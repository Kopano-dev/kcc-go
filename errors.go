/*
 * Copyright 2017 Kopano and its licensors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, version 3,
 * as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package kcc

import (
	"fmt"
)

// KCError is an error as returned by Kopano core.
type KCError uint64

func (err KCError) Error() string {
	return fmt.Sprintf("%s (KC:0x%x)", KCErrorText(err), uint64(err))
}

// Kopano Core error codes as defined in common/include/kopano/kcodes.h.
const (
	KCERR_NONE            = iota
	KCERR_UNKNOWN KCError = (1 << 31) | iota
	KCERR_NOT_FOUND
	KCERR_NO_ACCESS
	KCERR_NETWORK_ERROR
	KCERR_SERVER_NOT_RESPONDING
	KCERR_INVALID_TYPE
	KCERR_DATABASE_ERROR
	KCERR_COLLISION
	KCERR_LOGON_FAILED
	KCERR_HAS_MESSAGES
	KCERR_HAS_FOLDERS
	KCERR_HAS_RECIPIENTS
	KCERR_HAS_ATTACHMENTS
	KCERR_NOT_ENOUGH_MEMORY
	KCERR_TOO_COMPLEX
	KCERR_END_OF_SESSION
	KCWARN_CALL_KEEPALIVE
	KCERR_UNABLE_TO_ABORT
	KCERR_NOT_IN_QUEUE
	KCERR_INVALID_PARAMETER
	KCWARN_PARTIAL_COMPLETION
	KCERR_INVALID_ENTRYID
	KCERR_BAD_VALUE
	KCERR_NO_SUPPORT
	KCERR_TOO_BIG
	KCWARN_POSITION_CHANGED
	KCERR_FOLDER_CYCLE
	KCERR_STORE_FULL
	KCERR_PLUGIN_ERROR
	KCERR_UNKNOWN_OBJECT
	KCERR_NOT_IMPLEMENTED
	KCERR_DATABASE_NOT_FOUND
	KCERR_INVALID_VERSION
	KCERR_UNKNOWN_DATABASE
	KCERR_NOT_INITIALIZED
	KCERR_CALL_FAILED
	KCERR_SSO_CONTINUE
	KCERR_TIMEOUT
	KCERR_INVALID_BOOKMARK
	KCERR_UNABLE_TO_COMPLETE
	KCERR_UNKNOWN_INSTANCE_ID
	KCERR_IGNORE_ME
	KCERR_BUSY
	KCERR_OBJECT_DELETED
	KCERR_USER_CANCEL
	KCERR_UNKNOWN_FLAGS
	KCERR_SUBMITTED
)

// KCSuccess defines success response as returned by Kopano core.
const KCSuccess = KCERR_NONE

var kcErrorText = map[KCError]string{
	KCERR_UNKNOWN:               "Unknown",
	KCERR_NOT_FOUND:             "Not Found",
	KCERR_NO_ACCESS:             "No Access",
	KCERR_NETWORK_ERROR:         "Network Error",
	KCERR_SERVER_NOT_RESPONDING: "Server Not Responding",
	KCERR_INVALID_TYPE:          "Invalid Type",
	KCERR_DATABASE_ERROR:        "Database Erorr",
	KCERR_LOGON_FAILED:          "Logon Failed",
	KCERR_NOT_ENOUGH_MEMORY:     "Not Enough Memory",
	KCERR_END_OF_SESSION:        "End Of Session",
	KCERR_TIMEOUT:               "Timeout",
}

// KCErrorText returns a text for the KC error. It returns the empty string if
// the code is unknown.
func KCErrorText(code KCError) string {
	return kcErrorText[code]
}
