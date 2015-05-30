package bintray

const (
	libraryId      = "go-bintray"
	libraryVersion = "0.1"
	/*
		From Bintray docs:
		The latest version of the API will always be available at:
		https://api.bintray.com
		A specific version can be accessed at:
		https://bintray.com/api/v1
	*/
	defaultBaseURL      = "https://api.bintray.com/"
	userAgent           = libraryId + "/" + libraryVersion
	defaultDownloadHost = "https://dl.bintray.com/"
)
