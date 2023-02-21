package dal

import (
	"github.com/swordandtea/lets-habit-server/biz/response"
	"github.com/swordandtea/lets-habit-server/util"
	"gorm.io/gorm"
	"time"
)

type HabitCheckRecord struct {
	ID               uint64    `json:"id"`
	HabitID          uint64    `json:"habit_id"`
	UID              UID       `json:"uid"`
	CheckAt          time.Time `json:"-"`
	CheckAtTimeStamp string    `json:"check_at" gorm:"-"`
}

type habitCheckRecordDBHD struct{}

var HabitCheckRecordDBHD = &habitCheckRecordDBHD{}

func postProcessHabitCheckRecordField(records []*HabitCheckRecord) {
	for _, r := range records {
		r.CheckAtTimeStamp = util.GetCNTimeString(&r.CheckAt)
	}
}

func (hd *habitCheckRecordDBHD) Add(db *gorm.DB, r *HabitCheckRecord) response.SError {
	err := db.Create(r).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "add one habit check record fail")
	}
	postProcessHabitCheckRecordField([]*HabitCheckRecord{r})
	return nil
}

func (hd *habitCheckRecordDBHD) AddMulti(db *gorm.DB, rs []*HabitCheckRecord) response.SError {
	err := db.CreateInBatches(rs, 10).Error
	if err != nil {
		return response.ErrroCode_InternalUnknownError.Wrap(err, "add multi habit check record fail")
	}
	postProcessHabitCheckRecordField(rs)
	return nil
}
