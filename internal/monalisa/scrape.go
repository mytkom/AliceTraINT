package monalisa

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const MC_ANCH_PASS_NAME_COL_ID = 5
const MC_TAG_TARGET_STR = "alicemctag"
const MC_ANCH_TARGET_STR = "alicemcaprod"

type HyperloopRunList struct {
	RunToTags map[uint64][]string
	TagToRuns map[string][]uint64
	TagToIsMC map[string]bool
}

type hyperloopRunListRawEntry struct {
	EntryType          string `json:"type"`
	RunListStringified string `json:"runlist"`
	PeriodTag          string `json:"period"`
}

func (rl *HyperloopRunList) UnmarshalJSON(raw []byte) error {
	var rawEntries []hyperloopRunListRawEntry
	err := json.Unmarshal(raw, &rawEntries)
	if err != nil {
		return err
	}

	rl.TagToIsMC = make(map[string]bool, len(rawEntries))
	rl.TagToRuns = make(map[string][]uint64, len(rawEntries))
	rl.RunToTags = make(map[uint64][]string, len(rawEntries))

	for _, e := range rawEntries {
		for rs := range strings.SplitSeq(e.RunListStringified, ",") {
			rn, err := strconv.ParseUint(rs, 10, 64)
			if err != nil {
				return err
			}

			// check if run is present before appending
			if !slices.Contains(rl.TagToRuns[e.PeriodTag], rn) {
				rl.TagToRuns[e.PeriodTag] = append(rl.TagToRuns[e.PeriodTag], rn)
			}

			// check if tag is present before appending
			if !slices.Contains(rl.RunToTags[rn], e.PeriodTag) {
				rl.RunToTags[rn] = append(rl.RunToTags[rn], e.PeriodTag)
			}

			rl.TagToIsMC[e.PeriodTag] = (e.EntryType == "MC")
		}
	}

	return nil
}

type MonalisaClient struct {
	baseURL                   string
	cert                      tls.Certificate
	runListExpireAfterMinutes int64
	runList                   *HyperloopRunList
	runListObtainedAt         time.Time
}

func NewMonalisaClient(keyPath, certPath, baseUrl string) *MonalisaClient {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		log.Fatalf("cannot create MonalisaCert: %s", err.Error())
	}
	return &MonalisaClient{
		baseURL:                   baseUrl,
		cert:                      cert,
		runListExpireAfterMinutes: 1440, // 24 Hours
	}
}

type MCRow struct {
	MCTag         string
	AnchorProdTag string
	PassName      string
}

func (c *MonalisaClient) GetMCRow(mcPeriod string) (*MCRow, error) {
	url := fmt.Sprintf("%s/MC/?details=0&prodName=%s$", c.baseURL, mcPeriod)

	body := &bytes.Buffer{}

	request, err := http.NewRequest("GET", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{c.cert},
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upload failed with status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	mcTag := doc.Find(fmt.Sprintf(".link[target=%s]", MC_TAG_TARGET_STR)).
		First().
		Text()

	anchorProd := doc.Find(fmt.Sprintf(".link[target=%s]", MC_ANCH_TARGET_STR)).
		First().
		Text()

	passName := doc.Find("tbody tr.table_row td").
		Get(MC_ANCH_PASS_NAME_COL_ID).
		FirstChild.
		Data

	// trim string
	passName = strings.Trim(passName, " \n\t")

	// debug
	// fmt.Printf("MC Tag: %s, Anchor prod: %s, pass name: %s", mcTag, anchorProd, passName)

	return &MCRow{
		AnchorProdTag: anchorProd,
		MCTag:         mcTag,
		PassName:      passName,
	}, nil
}

func (c *MonalisaClient) GetRunList() (*HyperloopRunList, error) {
	expireTimestamp := int64(c.runListObtainedAt.Unix()) + (c.runListExpireAfterMinutes * int64(time.Minute))
	if c.runList != nil &&
		time.Now().Unix() <= expireTimestamp {
		return c.runList, nil
	}

	url := fmt.Sprintf("%s/alihyperloop-data/runlist/list-runlist.jsp?lists=runlists", c.baseURL)

	body := &bytes.Buffer{}
	req, err := http.NewRequest("GET", url, body)
	if err != nil {
		return nil, fmt.Errorf("problem with obtaining run list: %s", err.Error())
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{c.cert},
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get run list from hyperloop page: %w", err)
	}

	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get run list failed with status code %d", resp.StatusCode)
	}

	rawJson := bytes.Buffer{}
	_, err = rawJson.ReadFrom(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body of response failed %s", err.Error())
	}

	var runList HyperloopRunList
	err = json.Unmarshal(rawJson.Bytes(), &runList)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling response from JSON failed %s", err.Error())
	}

	c.runList = &runList
	c.runListObtainedAt = time.Now()

	return c.runList, nil
}
