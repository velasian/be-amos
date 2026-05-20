package importdata

import (
	"time"
)

type EmployeeStaging struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	ImportBatchID string `gorm:"index;size:50" json:"import_batch_id"`
	Status        string `gorm:"index;size:20;default:'PENDING'" json:"status"`
	ErrorMessage  string `gorm:"type:text" json:"error_message"`

	// Core Data (Raw)
	NRP              string `gorm:"size:50" json:"nrp"`
	Nama             string `gorm:"size:100" json:"nama"`
	Email            string `gorm:"size:100" json:"email"`
	Password         string `gorm:"size:100" json:"password"`
	JabatanRaw       string `gorm:"size:100" json:"jabatan_raw"`
	DepartemenRaw    string `gorm:"size:100" json:"departemen_raw"`
	JobSiteRaw       string `gorm:"size:100" json:"jobsite_raw"`
	StatusKaryawan   string `gorm:"size:50" json:"status_karyawan"`
	TanggalBergabung string `gorm:"size:20" json:"tanggal_bergabung"`

	// Details (Raw)
	TempatLahir           string `json:"tempat_lahir"`
	TanggalLahir          string `json:"tanggal_lahir"`
	JenisKelamin          string `json:"jenis_kelamin"`
	Agama                 string `json:"agama"`
	StatusPernikahan      string `json:"status_pernikahan"`
	GolonganDarah         string `json:"golongan_darah"`
	NIK                   string `json:"nik"`
	AlamatKTP             string `json:"alamat_ktp"`
	AlamatDomisili        string `json:"alamat_domisili"`
	NoHP                  string `json:"no_hp"`
	NoHPKeluarga          string `json:"no_hp_keluarga"`
	NamaKeluarga          string `json:"nama_keluarga"`
	HubunganKeluarga      string `json:"hubungan_keluarga"`
	NamaIbuKandung        string `json:"nama_ibu_kandung"`
	NoBPJSKesehatan       string `json:"no_bpjs_kesehatan"`
	NoBPJSKetenagakerjaan string `json:"no_bpjs_ketenagakerjaan"`
	NoNPWP                string `json:"no_npwp"`
	NamaBank              string `json:"nama_bank"`
	NoRekening            string `json:"no_rekening"`
	PemilikRekening       string `json:"pemilik_rekening"`
	UkuranBaju            string `json:"ukuran_baju"`
	UkuranSepatu          string `json:"ukuran_sepatu"`
	UkuranCelana          string `json:"ukuran_celana"`
	TinggiBadan           string `json:"tinggi_badan"`
	BeratBadan            string `json:"berat_badan"`
	TanggalPernikahan     string `json:"tanggal_pernikahan"`

	// Contract Data
	TipeKontrakRaw        string `json:"tipe_kontrak_raw"`
	NoSK                  string `json:"no_sk"`
	TanggalMulaiKontrak   string `json:"tanggal_mulai_kontrak"`
	TanggalSelesaiKontrak string `json:"tanggal_selesai_kontrak"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
