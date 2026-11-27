package repository

import (
	"strings"
	"time"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TrainingTaskRepository interface {
	Create(trainingTask *models.TrainingTask) error
	GetByID(id uint) (*models.TrainingTask, error)
	GetAll() ([]models.TrainingTask, error)
	GetAllUser(userId uint) ([]models.TrainingTask, error)
	List(query TrainingTaskListQuery) (TrainingTaskPage, error)
	DatasetChoices(search string, selected uint) ([]TrainingDatasetChoice, error)
	GetFirstQueued() (*models.TrainingTask, error)
	Update(trainingTask *models.TrainingTask) error
	Restart(userId uint, id uint) error
	Delete(userId uint, id uint) error
}

const TrainingTaskPageSize = 25

type TrainingTaskListQuery struct {
	DatasetID                  uint
	Name, Author               string
	Status                     *models.TrainingTaskStatus
	CreatedFrom, CreatedBefore time.Time
	Sort                       string
	Page                       int
}

type TrainingTaskPage struct {
	Tasks   []models.TrainingTask
	HasNext bool
}

type TrainingDatasetChoice struct {
	ID   uint   `json:"id,string"`
	Name string `json:"name"`
}

type trainingTaskRepository struct {
	db *gorm.DB
}

func NewTrainingTaskRepository(db *gorm.DB) TrainingTaskRepository {
	return &trainingTaskRepository{db: db}
}

func (r *trainingTaskRepository) Create(trainingTask *models.TrainingTask) error {
	return r.db.Create(trainingTask).Error
}

func (r *trainingTaskRepository) withDependencies() *gorm.DB {
	return r.db.
		Preload("TrainingDataset", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		}).
		Joins("User")
}

func (r *trainingTaskRepository) GetByID(id uint) (*models.TrainingTask, error) {
	var trainingTask models.TrainingTask
	if err := r.withDependencies().First(&trainingTask, id).Error; err != nil {
		return nil, err
	}
	return &trainingTask, nil
}

func (r *trainingTaskRepository) GetFirstQueued() (*models.TrainingTask, error) {
	var trainingTask models.TrainingTask
	if err := r.withDependencies().Where("\"status\" = ?", models.Queued).Order("\"training_tasks\".\"created_at\" asc").First(&trainingTask).Error; err != nil {
		return nil, err
	}

	return &trainingTask, nil
}

func (r *trainingTaskRepository) GetAll() ([]models.TrainingTask, error) {
	var trainingTasks []models.TrainingTask
	if err := r.withDependencies().Order("\"training_tasks\".\"created_at\" desc").Find(&trainingTasks).Error; err != nil {
		return nil, err
	}
	return trainingTasks, nil
}

func (r *trainingTaskRepository) GetAllUser(userId uint) ([]models.TrainingTask, error) {
	var trainingTasks []models.TrainingTask
	if err := r.withDependencies().Order("\"training_tasks\".\"created_at\" desc").Find(&trainingTasks, r.db.Where(&models.TrainingTask{UserId: userId})).Error; err != nil {
		return nil, err
	}
	return trainingTasks, nil
}

func (r *trainingTaskRepository) List(query TrainingTaskListQuery) (TrainingTaskPage, error) {
	db := r.db.Model(&models.TrainingTask{}).
		Select("training_tasks.id, training_tasks.name, training_tasks.created_at, training_tasks.updated_at, training_tasks.status, training_tasks.training_dataset_id").
		Joins("User").
		Preload("TrainingDataset", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped().Select("id", "name", "deleted_at")
		})

	if query.DatasetID != 0 {
		db = db.Where("training_tasks.training_dataset_id = ?", query.DatasetID)
	}
	if query.Name != "" {
		db = db.Where("LOWER(training_tasks.name) LIKE LOWER(?) ESCAPE '!'", taskListPattern(query.Name))
	}
	if query.Author != "" {
		pattern := taskListPattern(query.Author)
		db = db.Where(`(LOWER("User".first_name || ' ' || "User".family_name) LIKE LOWER(?) ESCAPE '!' OR LOWER("User".username) LIKE LOWER(?) ESCAPE '!')`, pattern, pattern)
	}
	if query.Status != nil {
		db = db.Where("training_tasks.status = ?", *query.Status)
	}
	if !query.CreatedFrom.IsZero() {
		db = db.Where("training_tasks.created_at >= ?", query.CreatedFrom)
	}
	if !query.CreatedBefore.IsZero() {
		db = db.Where("training_tasks.created_at < ?", query.CreatedBefore)
	}

	column, desc := "created_at", true
	switch query.Sort {
	case "oldest":
		desc = false
	case "updated":
		column = "updated_at"
	case "name-asc", "name-desc":
		column, desc = "name", query.Sort == "name-desc"
	}
	offset := (max(query.Page, 1) - 1) * TrainingTaskPageSize
	var page TrainingTaskPage
	err := db.Order(clause.OrderByColumn{Column: clause.Column{Table: "training_tasks", Name: column}, Desc: desc}).
		Order(clause.OrderByColumn{Column: clause.Column{Table: "training_tasks", Name: "id"}, Desc: desc}).
		Offset(offset).Limit(TrainingTaskPageSize + 1).Find(&page.Tasks).Error
	if len(page.Tasks) > TrainingTaskPageSize {
		page.Tasks = page.Tasks[:TrainingTaskPageSize]
		page.HasNext = true
	}
	return page, err
}

func taskListPattern(value string) string {
	return "%" + strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(value) + "%"
}

func (r *trainingTaskRepository) DatasetChoices(search string, selected uint) ([]TrainingDatasetChoice, error) {
	// Include archived datasets while a visible task still references them.
	db := r.db.Table("training_datasets").Select("id", "name").
		Where(`EXISTS (SELECT 1 FROM training_tasks
            WHERE training_tasks.training_dataset_id = training_datasets.id
            AND training_tasks.deleted_at IS NULL)`)

	choices := make([]TrainingDatasetChoice, 0)
	if selected != 0 {
		if err := db.Session(&gorm.Session{}).Where("id = ?", selected).Find(&choices).Error; err != nil {
			return nil, err
		}
	}
	var matches []TrainingDatasetChoice
	err := db.Where("id <> ? AND LOWER(name) LIKE LOWER(?) ESCAPE '!'", selected, taskListPattern(search)).
		Order("LOWER(name), id").Limit(TrainingTaskPageSize - len(choices)).Find(&matches).Error
	return append(choices, matches...), err
}

func (r *trainingTaskRepository) Update(trainingTask *models.TrainingTask) error {
	return r.db.Save(trainingTask).Error
}

func (r *trainingTaskRepository) Restart(userId uint, id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.TrainingTask{}).
			Where("id = ? AND user_id = ? AND status = ?", id, userId, models.Failed).
			Updates(map[string]interface{}{"status": models.Queued, "training_machine_id": nil})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return tx.Where("training_task_id = ? AND type IN ?", id, []models.TrainingTaskResultType{models.Image, models.Onnx}).
			Delete(&models.TrainingTaskResult{}).Error
	})
}

func (r *trainingTaskRepository) Delete(userId uint, id uint) error {
	return r.db.Where("\"user_id\" = ?", userId).Delete(&models.TrainingTask{}, id).Error
}

type MockTrainingTaskRepository struct {
	mock.Mock
}

func NewMockTrainingTaskRepository() *MockTrainingTaskRepository {
	return &MockTrainingTaskRepository{}
}

func (m *MockTrainingTaskRepository) Create(trainingTask *models.TrainingTask) error {
	args := m.Called(trainingTask)
	return args.Error(0)
}

func (m *MockTrainingTaskRepository) GetAll() ([]models.TrainingTask, error) {
	args := m.Called()

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]models.TrainingTask), args.Error(1)
}

func (m *MockTrainingTaskRepository) GetAllUser(userId uint) ([]models.TrainingTask, error) {
	args := m.Called(userId)

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]models.TrainingTask), args.Error(1)
}

func (m *MockTrainingTaskRepository) List(query TrainingTaskListQuery) (TrainingTaskPage, error) {
	args := m.Called(query)
	return args.Get(0).(TrainingTaskPage), args.Error(1)
}

func (m *MockTrainingTaskRepository) DatasetChoices(search string, selected uint) ([]TrainingDatasetChoice, error) {
	args := m.Called(search, selected)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]TrainingDatasetChoice), nil
}

func (m *MockTrainingTaskRepository) GetByID(id uint) (*models.TrainingTask, error) {
	args := m.Called(id)

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.TrainingTask), args.Error(1)
}
func (m *MockTrainingTaskRepository) GetFirstQueued() (*models.TrainingTask, error) {
	args := m.Called()

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.TrainingTask), args.Error(1)
}

func (m *MockTrainingTaskRepository) Update(trainingTask *models.TrainingTask) error {
	args := m.Called(trainingTask)
	return args.Error(0)
}

func (m *MockTrainingTaskRepository) Restart(userId uint, id uint) error {
	args := m.Called(userId, id)
	return args.Error(0)
}

func (m *MockTrainingTaskRepository) Delete(userId uint, id uint) error {
	args := m.Called(userId, id)
	return args.Error(0)
}
