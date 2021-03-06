// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package consensus

import (
	// noops      "github.com/monax/burrow/consensus/noops"
	tendermint "github.com/monax/burrow/consensus/tendermint"
)

//------------------------------------------------------------------------------
// Helper functions

func AssertValidConsensusModule(name, minorVersionString string) bool {
	switch name {
	case "noops":
		// noops should not have any external interfaces that can change
		// over iterations
		return true
	case "tendermint":
		return minorVersionString == tendermint.GetTendermintVersion().GetMinorVersionString()
	case "bigchaindb":
		// TODO: [ben] implement BigchainDB as consensus engine
		return false
	}
	return false
}
