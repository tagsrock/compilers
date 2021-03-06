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

package types

import (
	"testing"

	"time"

	consensus_types "github.com/monax/burrow/consensus/types"
	"github.com/tendermint/go-wire"
	tendermint_types "github.com/tendermint/tendermint/types"
)

func TestResultDumpConsensusState(t *testing.T) {
	result := ResultDumpConsensusState{
		ConsensusState: &consensus_types.ConsensusState{
			Height:     3,
			Round:      1,
			Step:       uint8(1),
			StartTime:  time.Now().Add(-time.Second * 100),
			CommitTime: time.Now().Add(-time.Second * 10),
			Validators: []consensus_types.Validator{
				&consensus_types.TendermintValidator{},
			},
			Proposal: &tendermint_types.Proposal{},
		},
		PeerConsensusStates: []*ResultPeerConsensusState{
			{
				PeerKey:            "Foo",
				PeerConsensusState: "Bar",
			},
		},
	}
	wire.JSONBytes(result)
}
