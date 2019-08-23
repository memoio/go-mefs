package ipfs

// CurrentCommit is the current git commit, this is set as a ldflag in the Makefile
var CurrentCommit string

// CurrentVersionNumber is the current application's version literal
const CurrentVersionNumber = "v0.2.0"

const ApiVersion = "/go-mefs/" + CurrentVersionNumber + "/"
