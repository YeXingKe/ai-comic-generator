package store

import (
	"encoding/json"

	"github.com/ai-comic-generator/server/internal/model"
	"gorm.io/gorm"
)

// CustomComicStore 自定义创作任务数据访问
type CustomComicStore struct {
	db *gorm.DB
}

func NewCustomComicStore(db *gorm.DB) *CustomComicStore {
	return &CustomComicStore{db: db}
}

func (s *CustomComicStore) Create(comic *model.CustomComic) error {
	return s.db.Create(comic).Error
}

func (s *CustomComicStore) GetByTaskID(taskID string) (*model.CustomComic, error) {
	var comic model.CustomComic
	err := s.db.Where("taskId = ? AND isDelete = 0", taskID).First(&comic).Error
	if err != nil {
		return nil, err
	}
	return &comic, nil
}

func (s *CustomComicStore) UpdateStatus(taskID, status string, errorMessage *string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if errorMessage != nil {
		updates["errorMessage"] = *errorMessage
	}
	return s.db.Model(&model.CustomComic{}).
		Where("taskId = ? AND isDelete = 0", taskID).
		Updates(updates).Error
}

func (s *CustomComicStore) SavePanelImages(taskID string, panels []model.PanelImageResult) error {
	b, err := json.Marshal(panels)
	if err != nil {
		return err
	}
	return s.db.Model(&model.CustomComic{}).
		Where("taskId = ? AND isDelete = 0", taskID).
		Update("panelImages", string(b)).Error
}

func (s *CustomComicStore) MarkCompleted(taskID string, panels []model.PanelImageResult) error {
	b, err := json.Marshal(panels)
	if err != nil {
		return err
	}
	return s.db.Model(&model.CustomComic{}).
		Where("taskId = ? AND isDelete = 0", taskID).
		Updates(map[string]interface{}{
			"status":      model.CustomComicStatusCompleted,
			"panelImages": string(b),
		}).Error
}

func (s *CustomComicStore) MarkFailed(taskID, errMsg string) error {
	return s.db.Model(&model.CustomComic{}).
		Where("taskId = ? AND isDelete = 0", taskID).
		Updates(map[string]interface{}{
			"status":       model.CustomComicStatusFailed,
			"errorMessage": errMsg,
		}).Error
}

// ListByPage 分页查询自定义创作任务
func (s *CustomComicStore) ListByPage(req *model.QueryCustomComicRequest) (*model.CustomComicPageResult, error) {
	q := s.db.Model(&model.CustomComic{}).Where("isDelete = 0")
	if req.UserID != nil {
		q = q.Where("userId = ?", *req.UserID)
	}
	if req.Status != nil && *req.Status != "" {
		q = q.Where("status = ?", *req.Status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	pageNum, pageSize := req.PageNum, req.PageSize
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	var rows []model.CustomComic
	err := q.Order("createTime DESC").
		Offset(int((pageNum - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	records := make([]model.CustomComicInfo, 0, len(rows))
	for i := range rows {
		if info := rows[i].ToCustomComicInfo(); info != nil {
			records = append(records, *info)
		}
	}
	return &model.CustomComicPageResult{
		Total: total, Records: records, PageNum: pageNum, PageSize: pageSize,
	}, nil
}
