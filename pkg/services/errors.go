/*
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package services

import (
	"fmt"
)

// ErrInvalidAuthorizedKeyFormat occurs when an authorized key in the SIP CR does not meet the expected format.
type ErrInvalidAuthorizedKeyFormat struct {
	SSHErr string
	Key    string
}

func (e ErrInvalidAuthorizedKeyFormat) Error() string {
	return fmt.Sprintf("encountered invalid Authorized Key: %s. The invalid key is %s", e.SSHErr, e.Key)
}

// ErrMalformedRedfishAddress occurs when a Redfish address does not meet the expected format.
type ErrMalformedRedfishAddress struct {
	Address string
}

func (e ErrMalformedRedfishAddress) Error() string {
	return fmt.Sprintf("invalid Redfish BMC address %s", e.Address)
}
