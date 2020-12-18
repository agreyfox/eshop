package paypal

//LiveIPNEndpoint contains the notification verification URL
const LiveIPNEndpoint = "https://www.paypal.com/cgi-bin/webscr"

//SandboxIPNEndpoint is the Sandbox notification verification URL
const SandboxIPNEndpoint = "https://ipnpb.sandbox.paypal.com/cgi-bin/webscr"

var (
	IPNEndpoint = SandboxIPNEndpoint
)

func getEndpoint(testIPN bool) string {
	/* 	if testIPN {
	   		return SandboxIPNEndpoint
	   	}
		   return LiveIPNEndpoint */
	return IPNEndpoint
}
