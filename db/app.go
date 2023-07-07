package db

var (
	AppVersion     = "0.0.0.unknown"
	AppName        = "dae-wing"
	AppDescription = ""
)

func init() {
	if AppDescription == "" {
		AppDescription = AppName + " is a integration solution of dae, API and UI."
	}
}
