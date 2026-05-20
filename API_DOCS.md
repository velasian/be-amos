# AMOS Backend API Documentation

**Base URL**: `/api/v1`

## Authentication (`/auth`)
| Method | Endpoint | Description | Auth Required | Role |
|---|---|---|---|---|
| POST | `/auth/login` | Authenticate user & get tokens | No | Any |
| POST | `/auth/refresh` | Refresh access token using cookie | No | Any |
| POST | `/auth/forgot-password` | Request password reset OTP | No | Any |
| POST | `/auth/verify-otp` | Verify the OTP sent to email | No | Any |
| POST | `/auth/reset-password` | Set new password after OTP | No | Any |
| POST | `/auth/logout` | Logout & clear refresh token | Yes | Any |
| POST | `/auth/fcm-token` | Save Firebase Cloud Messaging token | Yes | Any |

## Masters (`/masters`)
| Method | Endpoint | Description | Auth Required | Role |
|---|---|---|---|---|
| GET | `/masters/departments` | Get all departments | Yes | Any |
| POST | `/masters/departments` | Create new department | Yes | `admin_hr` |
| PUT | `/masters/departments/:id` | Update department | Yes | `admin_hr` |
| DELETE | `/masters/departments/:id` | Delete department | Yes | `admin_hr` |
| GET | `/masters/positions` | Get all positions | Yes | Any |
| POST | `/masters/positions` | Create new position | Yes | `admin_hr` |
| PUT | `/masters/positions/:id` | Update position | Yes | `admin_hr` |
| DELETE | `/masters/positions/:id` | Delete position | Yes | `admin_hr` |
| GET | `/masters/job-sites` | Get all job sites | Yes | Any |
| POST | `/masters/job-sites` | Create new job site | Yes | `admin_hr` |
| PUT | `/masters/job-sites/:id` | Update job site | Yes | `admin_hr` |
| DELETE | `/masters/job-sites/:id` | Delete job site | Yes | `admin_hr` |
| GET | `/masters/contract-types` | Get all contract types | Yes | Any |
| POST | `/masters/contract-types` | Create contract type | Yes | `admin_hr` |
| PUT | `/masters/contract-types/:id`| Update contract type | Yes | `admin_hr` |
| DELETE | `/masters/contract-types/:id`| Delete contract type | Yes | `admin_hr` |

## Employees (`/employees`)
| Method | Endpoint | Description | Auth Required | Role |
|---|---|---|---|---|
| GET | `/employees/me` | Get currently logged-in user profile | Yes | Any |
| PUT | `/employees/me` | Update logged-in user profile | Yes | Any |
| GET | `/employees` | Get all employees (paginated/filterable) | Yes | `admin_hr` |
| POST | `/employees` | Create new employee | Yes | `admin_hr` |
| GET | `/employees/:id` | Get employee by ID | Yes | `admin_hr` |
| PUT | `/employees/:id` | Update employee data | Yes | `admin_hr` |
| DELETE | `/employees/:id` | Delete employee | Yes | `admin_hr` |
| GET | `/employees/export` | Export employees data to Excel | Yes | `admin_hr` |

## File Management (`/files`)
| Method | Endpoint | Description | Auth Required | Role |
|---|---|---|---|---|
| GET | `/files` | Get files by entity/module | Yes | Any |
| GET | `/files/:id/download` | Get file download pre-signed URL | Yes | Any |
| POST | `/files/upload` | Upload a file to MinIO | Yes | `admin_hr` |
| DELETE | `/files/:id` | Delete a file from MinIO | Yes | `admin_hr` |

## Employee Import Staging (`/import`)
| Method | Endpoint | Description | Auth Required | Role |
|---|---|---|---|---|
| GET | `/import/template` | Download Excel import template | Yes | `admin_hr` |
| POST | `/import/parse` | Upload & parse Excel, returns `batchId` | Yes | `admin_hr` |
| GET | `/import/staging/:batchId`| Get staging data for review | Yes | `admin_hr` |
| PATCH | `/import/staging/:id` | Update specific field in staging row | Yes | `admin_hr` |
| POST | `/import/commit` | Commit valid staging data to database | Yes | `admin_hr` |

## Notifications (`/notifications`)
| Method | Endpoint | Description | Auth Required | Role |
|---|---|---|---|---|
| GET | `/notifications` | Get user's notification inbox | Yes | Any |
| GET | `/notifications/unread-count`| Get unread notification count | Yes | Any |
| PATCH | `/notifications/:id/read` | Mark specific notification as read | Yes | Any |
| PATCH | `/notifications/read-all` | Mark all notifications as read | Yes | Any |
| POST | `/notifications/test` | Send test notification (Admin test) | Yes | `admin_hr` |

## IoT & Hardware (`/iot`)

Modul ini memfasilitasi komunikasi antara perangkat keras *scanner* (seperti **ESP32 + RFID/NFC Reader**) dengan sistem *backend* AMOS, serta fitur manajemen perangkat untuk admin HR.

**Keamanan Perangkat IoT:**
Semua *endpoint* yang diakses langsung oleh perangkat keras (mesin absensi) **TIDAK** menggunakan JWT Bearer. Sebagai gantinya, perangkat harus mengirimkan API Key rahasianya di dalam *header* `X-IoT-Key`.

### 1. Endpoint Mesin Absensi (Diakses oleh ESP32)

| Method | Endpoint | Auth Header | Deskripsi |
|---|---|---|---|
| **POST** | `/iot/scan` | `X-IoT-Key` | Digunakan saat karyawan melakukan *tap* kartu untuk presensi (Check In / Check Out).<br>**Payload JSON:** `{ "nfc_uid": "AB:CD:EF:12" }`<br>**Respon:** Berhasil mencatat kehadiran atau peringatan (seperti kartu belum terdaftar). |
| **POST** | `/iot/assign` | `X-IoT-Key` | Digunakan saat mesin absensi dalam mode "Pendaftaran Kartu" (*Registration Mode*). Mem- *broadcast* UID yang baru saja di-*tap* ke seluruh dasbor admin HR secara seketika (*Real-time*).<br>**Payload JSON:** `{ "nfc_uid": "AB:CD:EF:12" }` |

### 2. Endpoint Manajemen Perangkat (Diakses oleh Admin HR)

| Method | Endpoint | Auth Header | Deskripsi |
|---|---|---|---|
| **POST** | `/iot/devices` | `Authorization: Bearer` | Mendaftarkan mesin ESP32 baru ke sistem. <br>**Payload JSON:** `{ "name": "Scanner Pintu Depan", "job_site_id": 1 }`<br>**Respon Krusial:** Menghasilkan `api_key` (Contoh: `iot_a1b2c3d4...`) yang **HANYA DITAMPILKAN SEKALI**. Key ini harus di-*hardcode* atau diunggah ke mesin ESP32. |
| **GET** | `/iot/devices` | `Authorization: Bearer` | Melihat daftar semua mesin absensi yang terdaftar di seluruh *site* operasional. (API Key tidak akan dimunculkan di sini demi keamanan). |
| **GET** | `/iot/listen` | *SSE Stream* (Browser) | **Mekanisme Server-Sent Events (SSE)**.<br>Dasbor HR akan membuka koneksi *stream* ini. Jika ada yang menempelkan kartu kosong ke mesin (melalui `/iot/assign`), *backend* akan langsung mendorong *event* ke *endpoint* ini tanpa perlu *refresh* halaman. |
| **POST** | `/iot/assign-employee` | `Authorization: Bearer` | Menetapkan ( *assign* ) kartu NFC yang baru saja dipindai ke seorang karyawan tertentu di *database*.<br>**Payload JSON:** `{ "employee_id": 12, "nfc_uid": "AB:CD:EF:12" }` |

## Attendances (`/attendances`)

Modul ini menangani pelacakan sesi kehadiran, baik secara mandiri melalui aplikasi *mobile* maupun *monitoring* oleh HR. Semua *endpoint* membutuhkan **JWT Bearer**.

| Method | Endpoint | Query/Body | Deskripsi |
|---|---|---|---|
| **GET** | `/attendances/session` | - | Mengambil sesi absensi aktif (*pending*) untuk karyawan yang sedang *login*. Berguna untuk aplikasi *mobile* mengecek apakah *user* sudah *Clock In* hari ini dan menunggu *Clock Out*. |
| **GET** | `/attendances/me` | `?page=1`<br>`&limit=20`<br>`&start_date=2026-05-01`<br>`&end_date=2026-05-31` | Mengambil riwayat absensi pribadi dari pengguna yang sedang *login* (mendukung filter tanggal dan paginasi). |
| **POST** | `/attendances/verify` | **Content-Type**: `multipart/form-data`<br><br>Fields:<br>`session_id` (int)<br>`latitude` (float)<br>`longitude` (float)<br>`selfie` (file gambar) | Mengirim data *Clock In / Clock Out* melalui **Geofence Mobile App**. Sistem akan memverifikasi apakah koordinat GPS pengguna berada di dalam radius *Job Site*, dan menyimpan bukti foto *selfie*. |
| **GET** | `/attendances` | `?page=1`<br>`&search=budi`<br>`&department_id=2`<br>`&job_site_id=1` | *(Hanya `admin_hr`)*<br>Melihat daftar seluruh riwayat absensi di semua cabang operasional secara komprehensif dengan dukungan filter spesifik. |

## Reports (`/reports`)
| Method | Endpoint | Description | Auth Required | Role |
|---|---|---|---|---|
| GET | `/reports/stats` | Get dashboard statistics | Yes | `admin_hr` |
| GET | `/reports/export` | Export attendance reports to Excel | Yes | `admin_hr` |

---
**Note:** All endpoints under `/api/v1` except for `/auth/login`, `/auth/refresh`, and password reset APIs require a valid JWT Bearer token in the `Authorization` header. IoT endpoints use a custom API Key header `X-IoT-Key`.
