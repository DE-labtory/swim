/*
 * Copyright 2018 De-labtory
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package swim

type Config struct {
	MaxlocalCount int
}

type SWIM struct {

	// Swim Config
	config *Config

	// Currently connected memberList
	memberMap MemberMap
}

// Start SWIM protocol.
func (s *SWIM) Start() {

}

// Dial to the all peerAddresses and exchange memberList.
func (s *SWIM) Join(peerAddresses []string) error {
	return nil
}

// Gossip message to p2p network.
func (s *SWIM) Gossip(msg []byte) {

}

// Shutdown the running swim.
func (s *SWIM) ShutDown() {

}

// Failure Detection is performed for each` T`. (ref: https://github.com/DE-labtory/swim/edit/develop/docs/Docs.md)
//
// 1. SWIM randomly selects a member(j) in the memberMap and ping to the member(j).
//
// 2. SWIM waits for ack of the member(j) during the ack-timeout (time less than T).
//    End failure Detector if ack message arrives on ack-timeout.
//
// 3. SWIM selects k number of members from the memberMap and sends indirect-ping(request k members to ping the member(j)).
//    The nodes (that receive the indirect-ping) ping to the member(j) and ack when they receive ack from the member(j).
//
// 4. At the end of T, SWIM checks to see if ack was received from k members, and if there is no message,
//    The member(j) is judged to be failed, so check the member(j) as suspected or delete the member(j) from memberMap.
//
// ** When performing ping, ack, and indirect-ping in the above procedure, piggybackdata is sent together. **
func (s *SWIM) startFailureDetector() {

}

// Delegate interface to notify status change of the member
type MemberStatusChangeDelegate interface {
	deleteMember(id MemberID)
	addMember()
	updateMember()
}

// Delete member
func (s *SWIM) deleteMember(id MemberID) {

}

// Add new member
func (s *SWIM) addMember() {

}

// Update member
func (s *SWIM) updateMember() {

}
