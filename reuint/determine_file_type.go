package reuint

func GetMIME(fileType string) string {
	switch fileType {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".md":
		return "text/markdown"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".txt":
		return "text/plain"
	case ".zip":
		return "application/zip"
	case ".crt":
		return "application/x-x509-ca-cert"
	default:
		return "application/octet-stream"
	}
	return ""
}
