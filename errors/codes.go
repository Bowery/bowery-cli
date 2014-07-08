// Copyright 2013-2014 Bowery, Inc.
package errors

type Error struct {
	Code        string
	Title       string
	Description string
}

var errorCodes = []Error{
	Error{
		Code:  "0",
		Title: ErrNoDeveloper.Error(),
		Description: "The active developer is not logged into Bowery. To do so, simply run\n\n" +
			"$ bowery login\n\n" +
			"The current developer can be seen at any time by running\n\n" +
			"$ bowery info\n",
	},
	Error{
		Code:        "1",
		Title:       ErrDeveloperExists.Error(),
		Description: "Every developer is identified by a unique ID and a unique email address.",
	},
	Error{
		Code:  "2",
		Title: ErrNotConnected.Error(),
		Description: "In most cases if the application has never been `connected` the local\n" +
			"state and server state are not in sync. Running connect will provide\n" +
			"the neccessary info to execute a variety of commands.",
	},
	Error{
		Code:  "3",
		Title: ErrInvalidCommand.Error(),
		Description: "Bowery has a preset number of commands that can be executed.\n" +
			"To view them all, run\n\n" +
			"$ bowery help\n",
	},
	Error{
		Code:  "4",
		Title: ErrInvalidService.Error(),
		Description: "Commands that restart, save, destroy, etc. services require a valid\n" +
			"service name to be provided. Active services can be seen by running\n\n" +
			"$ bowery info\n\n" +
			"or looking in your bowery.json file.",
	},
	Error{
		Code:        "5",
		Title:       ErrInvalidLogin.Error(),
		Description: "The user attempting to login provided the wrong email and password combo.",
	},
	Error{
		Code:        "6",
		Title:       ErrTooManyLogins.Error(),
		Description: "Login attempts are limited to 5 tries.",
	},
	Error{
		Code:  "7",
		Title: ErrNoServices.Error(),
		Description: "Bowery applications are comprised of services. In order to use Bowery\n" +
			"there must be at least one service to connect and sync with.",
	},
	Error{
		Code:  "8",
		Title: ErrNoServicePaths.Error(),
		Description: "To sync files, a path must be specified within that service.\n" +
			"In some cases a path is not needed, e.g. a database. To update a path\n" +
			"specify a path in the bowery.json file.",
	},
	Error{
		Code:  "9",
		Title: ErrCantConnect.Error(),
		Description: "Trouble connecting to the Bowery API. Bowery requires a working\n" +
			"internet connection. If a connection is present, this means there is an error with\n" +
			"the Bowery API. The team at Bowery closely monitors the status of it's API.",
	},
	Error{
		Code:  "10",
		Title: ErrFailedRestart.Error(),
		Description: "Bowery failed to successfully restart a service. This can the be the\n" +
			"result of a bad request (invalid service) or a internal server error (Bowery's fault).\n" +
			"Make sure that `bowery connect` works as expected. If not, contact support.",
	},
	Error{
		Code:  "11",
		Title: ErrMismatchPass.Error(),
		Description: "The password provided, and it's confirmation do not match. Make sure\n" +
			"you properly enter your password both times.",
	},
	Error{
		Code:  "12",
		Title: ErrOutOfDate.Error(),
		Description: "A later version of Bowery is available. For the time being, it is required\n" +
			"that the latest cli is used.",
	},
	Error{
		Code:        "13",
		Title:       ErrVersionDownload.Error(),
		Description: "Check http://docs.bowery.io/#install for further installation instructions.",
	},
	Error{
		Code:  "14",
		Title: ErrSyncFailed.Error(),
		Description: "A system error occured within the service. If restarting does not resolve the\n" +
			"issue, try reconnecting or cleaning.",
	},
	Error{
		Code:  "15",
		Title: ErrImageExists.Error(),
		Description: "All Bowery images are identified by a unique name. When saving a service\n" +
			"a unique name must be provided. Run `bowery search 'name'` to verify if the\n" +
			"name has been taken.",
	},
	Error{
		Code:  "16",
		Title: ErrNoImageFound.Error(),
		Description: "All Bowery images are identified by a unique name. When creating a service\n" +
			"a valid image name must be provided. Run `bowery search 'name'` to find available images.",
	},
	Error{
		Code:  "17",
		Title: ErrIORedirection.Error(),
		Description: "IO redirection is not possible with Bowery ssh, this is because\n" +
			"it needs to be connected to a terminal to get the window size.",
	},
	Error{
		Code:  "18",
		Title: ErrPathNotFoundTmpl,
		Description: "In order to sync code between your local machine and Bowery, an existing\n" +
			"directory must be provided.",
	},
	Error{
		Code:  "19",
		Title: ErrInvalidJSONTmpl,
		Description: "Bowery applications are defined in a bowery.json file. In order for it to be\n" +
			"properly processed, it must be formatted as valid JSON.",
	},
	Error{
		Code:  "20",
		Title: ErrInvalidPortTmpl,
		Description: "Bowery allows you to specify additional ports to make available. By default\n" +
			"ports 80 (HTTP), 22 (SSH), and 3001 (reserved by Bowery) are created. The ports specified\n" +
			"must be valid integers in the standard range.",
	},
	Error{
		Code:  "21",
		Title: ErrUpdatePerm.Error(),
		Description: "Some installations may reside in write protected directories, so to update Bowery\n" +
			"in these directories administrator privileges are required.",
	},
	Error{
		Code:        "22",
		Title:       ErrInvalidConfigKey.Error(),
		Description: "The config key isn't valid, view `bowery help config` to get the list of valid keys.",
	},
	Error{
		Code:  "23",
		Title: ErrInvalidToken.Error(),
		Description: "The token for your developer is invalid. This may occur if the .boweryconf file has\n" +
			"been edited or if a login has occured somewhere else.\n\nTo correct this, just logout and log back in.",
	},
	Error{
		Code:  "24",
		Title: ErrOverCapacity.Error(),
		Description: "All of Bowery's resources are currently being used up. We have been notified and will\n" +
			"do our best to get more servers online immediately.",
	},
	Error{
		Code:  "25",
		Title: ErrContainerConnect.Error(),
		Description: "The container running the service to which you're trying to connect is down.\n" +
			"Running `bowery restart` with the name of the service should help.",
	},
	Error{
		Code:  "26",
		Title: ErrResetRequest.Error(),
		Description: "Something went wrong while trying to reset your password and send you an email.\n" +
			"Please try again.",
	},
}

func GetAll() []Error {
	return errorCodes
}

func Get(index int) (Error, error) {
	if index < 0 || index >= len(errorCodes) {
		return Error{}, Newf(ErrErrorsRange, len(errorCodes))
	}

	return errorCodes[index], nil
}
