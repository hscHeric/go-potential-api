package domain

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	AuthID    uint      `gorm:"uniqueIndex;not null" json:"auth_id"`
	FullName  string    `gorm:"not null" json:"full_name"`
	CPF       string    `gorm:"uniqueIndex;not null" json:"cpf"`
	BirthDate time.Time `gorm:"not null" json:"birth_date"`

	// JSONB fields
	Address    datatypes.JSON `gorm:"type:jsonb" json:"address"`
	Contact    datatypes.JSON `gorm:"type:jsonb" json:"contact"`
	Documents  datatypes.JSON `gorm:"type:jsonb" json:"documents"` // URLs dos documentos, se enviar arquivos
	ProfilePic string         `json:"profile_pic"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Address struct {
	Street     string `json:"street"`
	Number     string `json:"number"`
	Complement string `json:"complement,omitempty"`
	District   string `json:"district"`
	City       string `json:"city"`
	State      string `json:"state"`
	ZipCode    string `json:"zip_code"`
}

type Contact struct {
	Phone       string `json:"phone"`
	MobilePhone string `json:"mobile_phone,omitempty"`
	WhatsApp    string `json:"whatsapp,omitempty"`
}

type Documents struct {
	RG             string `json:"rg,omitempty"`
	CPFDoc         string `json:"cpf_doc,omitempty"`
	ProofOfAddress string `json:"proof_of_address,omitempty"`
}

type CompleteProfileRequest struct {
	FullName  string  `json:"full_name" binding:"required"`
	CPF       string  `json:"cpf" binding:"required"`
	BirthDate string  `json:"birth_date" binding:"required"` // YYYY-MM-DD
	Password  string  `json:"password" binding:"required,min=6"`
	Address   Address `json:"address" binding:"required"`
	Contact   Contact `json:"contact" binding:"required"`
}
