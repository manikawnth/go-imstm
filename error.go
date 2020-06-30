package imstm

import "strconv"

// returnCodes indicate IMS connect return codes
var returnCodes map[int]string = map[int]string{
	4:  "Exit request error message sent to client before socket termination",
	8:  "Error detected by IMS Connect and the socket is disconnected for IMS",
	12: "Error returned by IMS OTMA and the socket is disconnected for IMS",
	16: "Error returned by IMS OTMA when an OTMA sense code is returned in RSM",
	20: "Exit requests response message to HWSPWCH/PING request to be returned to client",
	24: "SCI error detected, socket is disconnected for IMS",
	28: "OM error detected, socket is disconnected for IMS",
	32: "IRM_TIMER value expired and socket is disconnected by IMS Connect",
	36: "A default IRM_TIMER value expired and socket is disconnected by IMS Connect",
	40: "IRM_TIMER value expired but the socket remains connected",
	44: "Datastore is no longer available",
}

// ReturnCode indicates the error code returned by IMS connect
// Errors could've been returned from OTMA or IMS
type ReturnCode int

// String returns the IMS connect response return code as string
func (rc ReturnCode) String() string {
	if str, ok := returnCodes[int(rc)]; ok {
		return str
	}
	return "Unknown: " + strconv.Itoa(int(rc))
}

// reasonCodes indicate IMS connect reason codes
var reasonCodes map[int]string = map[int]string{
	4:  "Input data exceeds buffer size",
	5:  "Negative length value",
	6:  "IRM- IMS request message length invalid",
	7:  "Total message length invalid",
	8:  "OTMA NAK with no sense code or RC",
	9:  "Contents of buffer invalid",
	10: "Output data exceeds buffer size",
	11: "Invalid unicode definition",
	12: "Invalid message, no data",
	16: "Do not know who client is",
	20: "OTMA segment length error",
	24: "FIC missing",
	28: "LIC missing",
	32: "Sequence number error",
	34: "Unable to locate context token",
	36: "Protocol error",
	40: "Security violation",
	44: "Message incomplete",
	48: "Incorrect message length",
	51: "Security failure: no OTMA security header",
	52: "Security failure: no security data in OTMA security header",
	53: "Security failure: no password in OTMA user data header",
	54: "Security failure: no user ID in OTMA security header",
	55: "Security failure: no password in OTMA user data and no user ID in OTMA security header",
	56: "Duplicate Client ID used; the client ID is currently in use",
	57: "Invalid token is being used: internal error",
	58: "Invalid client status: internal error",
	59: "Cancel Timer completed successfully",
	70: "Component not found",
	71: "Function not found",
	72: "The data store was not found, or communication with the data store was stopped using an IMS Connect command",
	73: "IMS Connect in shutdown",
	74: "The data store or IMSplex was in a stop or close process, or the IMS data store has shut down or disconnected",
	75: "Data store communication error",
	76: "The data store or IMSplex was stopped by command",
	77: "A data store or IMSplex communication error was sent to the pending client",
	78: "Security failure. RACF call failed, IMS Connect call failed. See IMS Connect error message on system console",
	79: "IMS Connect protocol error",
	80: "The IMSplex connection is not active",
	81: "IMS cancelled the Resume Tpipe request as a result of an ACKTO timeout",
	93: "Invalid commit mode of 1 specified on the RESUME TPIPE request",
	94: "Request",
	95: "Conversation",
	96: "Request and conversation",
	97: "Deallocate confirmed",
	98: "Deallocate abnormal termination",
	99: "Default reason code",
}

// ReasonCode indicates the IMS connect response reason code
// Errors could've been returned from OTMA or IMS
type ReasonCode int

// String returns the IMS connect response reason code as string
func (rc ReasonCode) String() string {
	if str, ok := returnCodes[int(rc)]; ok {
		return str
	}
	return "Unknown: " + strconv.Itoa(int(rc))
}
