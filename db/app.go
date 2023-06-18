package db

var (
	AppVersion     = ""
	AppName        = ""
	AppDescription = ""
)

func init() {
	if AppDescription == "" {
		AppDescription = AppName + " is a integration solution of dae, API and UI."
	}
}
