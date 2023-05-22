package polybft

import (
	"fmt"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
	"github.com/umbracle/fastrlp"
)

type ExtraHandlerLondon struct {
}

// MarshalRLPWith defines the marshal function implementation for Extra
func (e *ExtraHandlerLondon) MarshalRLPWith(extra *Extra, ar *fastrlp.Arena) *fastrlp.Value {
	if extra.Logger != nil {
		extra.Logger.Warn("MarshalRLPWith London style", "dummy1", extra.Dummy1)
	}

	vv := extraMarshalBaseFields(extra, ar, ar.NewArray())

	vv.Set(ar.NewString(extra.Dummy1))

	return vv
}

// UnmarshalRLPWith defines the unmarshal implementation for Extra
func (e *ExtraHandlerLondon) UnmarshalRLPWith(extra *Extra, v *fastrlp.Value) error {
	if extra.Logger != nil {
		extra.Logger.Warn("UnmarshalRLPWith London style")
	}

	elems, err := extraUnmarshalBaseFields(extra, v, 5)
	if err != nil {
		return err
	}

	extra.Dummy1, err = elems[4].GetString()
	if err != nil {
		return fmt.Errorf("dummy1 bytes: %w", err)
	}

	return nil
}

func (e *ExtraHandlerLondon) ValidateAdditional(extra *Extra, header *types.Header, logger hclog.Logger) error {
	logger.Warn("ValidateAdditional London style")

	if extra.Dummy1 == "" {
		return fmt.Errorf("dummy1 is empty for block %d", header.Number)
	}

	return nil
}

type ExtraHandlerAdditionalLondon struct {
}

func (i *ExtraHandlerAdditionalLondon) GetIbftExtraClean(extra *Extra) *Extra {
	return &Extra{
		BlockNumber: extra.BlockNumber,
		Parent:      extra.Parent,
		Validators:  extra.Validators,
		Checkpoint:  extra.Checkpoint,
		Committed:   &Signature{},
		Dummy1:      extra.Dummy1,
	}
}
