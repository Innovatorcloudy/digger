// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameProject = "projects"

// Project mapped from table <projects>
type Project struct {
	ID                 string         `gorm:"column:id;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt          time.Time      `gorm:"column:created_at;default:now()" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
	Name               string         `gorm:"column:name;not null" json:"name"`
	RepoID             string         `gorm:"column:repo_id;not null" json:"repo_id"`
	DriftEnabled       bool           `gorm:"column:drift_enabled" json:"drift_enabled"`
	DriftStatus        string         `gorm:"column:drift_status" json:"drift_status"`
	LatestDriftCheck   time.Time      `gorm:"column:latest_drift_check" json:"latest_drift_check"`
	DriftTerraformPlan string         `gorm:"column:drift_terraform_plan" json:"drift_terraform_plan"`
	ToUpdate           int32          `gorm:"column:to_update" json:"to_update"`
	ToChange           int32          `gorm:"column:to_change" json:"to_change"`
	ToDelete           int32          `gorm:"column:to_delete" json:"to_delete"`
	IsAcknowledged     bool           `gorm:"column:is_acknowledged;not null" json:"is_acknowledged"`
}

// TableName Project's table name
func (*Project) TableName() string {
	return TableNameProject
}
