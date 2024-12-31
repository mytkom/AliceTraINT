package jalien

import (
	"testing"
)

func TestParseLongFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *longFormatParsed
		wantErr bool
	}{
		{
			name:  "Valid file line",
			input: "-rw-r--r-- user group 12345 Jan 01 12:34 somefile.txt",
			want: &longFormatParsed{
				Permissions: "-rw-r--r--",
				Owner:       "user",
				Group:       "group",
				Size:        12345,
				Month:       "Jan",
				Day:         "01",
				Time:        "12:34",
				Name:        "somefile.txt",
				IsDir:       false,
			},
			wantErr: false,
		},
		{
			name:  "Valid directory line",
			input: "drwxr-xr-x user group 4096 Feb 15 08:00 somedir/",
			want: &longFormatParsed{
				Permissions: "drwxr-xr-x",
				Owner:       "user",
				Group:       "group",
				Size:        4096,
				Month:       "Feb",
				Day:         "15",
				Time:        "08:00",
				Name:        "somedir/",
				IsDir:       true,
			},
			wantErr: false,
		},
		{
			name:    "Invalid line format",
			input:   "-rw-r--r-- user group 12345",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLongFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLongFormat() error = %v, wantErr %t", err, tt.wantErr)
				return
			}
			if !tt.wantErr && *got != *tt.want {
				t.Errorf("parseLongFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}
