package modules

import (
	"github.com/NebulousLabs/Sia/types"
)

const (
	AcceptTermsResponse = "accept"
	HostDir             = "host"
)

// ContractTerms are the parameters agreed upon by a client and a host when
// forming a FileContract.
type ContractTerms struct {
	FileSize           uint64                // How large the file is.
	Duration           types.BlockHeight     // How long the file is to be stored.
	DurationStart      types.BlockHeight     // The block height that the storing starts (typically required to start immediately, unless it's a chained contract).
	WindowSize         types.BlockHeight     // How long the host has to submit a proof of storage.
	Price              types.Currency        // Client contribution towards payout each window
	Collateral         types.Currency        // Host contribution towards payout each window
	ValidProofOutputs  []types.SiacoinOutput // Where money goes if the storage proof is successful.
	MissedProofOutputs []types.SiacoinOutput // Where the money goes if the storage proof fails.
}

type HostInfo struct {
	HostSettings

	StorageRemaining int64
	NumContracts     int
}

type Host interface {
	// Announce announces the host on the blockchain.
	Announce(NetAddress) error

	// HostNotify will push a struct down the channel every time that an update
	// is received.
	HostNotify() <-chan struct{}

	// NegotiateContract is an RPC that enables a client to communicate with
	// the host to propose a contract.
	//
	// TODO: enhance this documentataion. For now, see the host package for a
	// reference implementation.
	//NegotiateContract(NetConn) error

	// RetrieveFile is an RPC that enables a client to download a file from
	// the host.
	//RetrieveFile(NetConn) error

	// SetConfig sets the hosting parameters of the host.
	SetSettings(HostSettings)

	// Settings is an RPC that returns the host's settings.
	//Settings(NetConn) error

	// Info returns info about the host, including its hosting parameters, the
	// amount of storage remaining, and the number of active contracts.
	Info() HostInfo
}
