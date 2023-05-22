package polybft

import (
	"fmt"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
	"github.com/umbracle/fastrlp"
)

type ExtraHandlerNikaragva struct {
}

// MarshalRLPWith defines the marshal function implementation for Extra
func (e *ExtraHandlerNikaragva) MarshalRLPWith(extra *Extra, ar *fastrlp.Arena) *fastrlp.Value {
	if extra.Logger != nil {
		extra.Logger.Warn("MarshalRLPWith Nikaragva style")
	}

	vv := extraMarshalBaseFields(extra, ar, ar.NewArray())

	vv.Set(ar.NewString(extra.Dummy1))
	vv.Set(ar.NewString(extra.Dummy2))

	return vv
}

// UnmarshalRLPWith defines the unmarshal implementation for Extra
func (e *ExtraHandlerNikaragva) UnmarshalRLPWith(extra *Extra, v *fastrlp.Value) error {
	if extra.Logger != nil {
		extra.Logger.Warn("UnmarshalRLPWith Nikaragva style")
	}

	elems, err := extraUnmarshalBaseFields(extra, v, 6)
	if err != nil {
		return err
	}

	extra.Dummy1, err = elems[4].GetString()
	if err != nil {
		return fmt.Errorf("dummy1 bytes: %w", err)
	}

	extra.Dummy2, err = elems[5].GetString()
	if err != nil {
		return fmt.Errorf("dummy2 bytes: %w", err)
	}

	return nil
}

func (e *ExtraHandlerNikaragva) ValidateAdditional(extra *Extra, header *types.Header, logger hclog.Logger) error {
	logger.Warn("ValidateAdditional Nikaragva style")

	if extra.Dummy1 == "" {
		return fmt.Errorf("dummy1 is empty for block %d", header.Number)
	}

	if extra.Dummy2 == "" {
		return fmt.Errorf("dummy2 is empty for block %d", header.Number)
	}

	return nil
}

type ExtraHandlerAdditionalNikaragva struct {
}

func (i *ExtraHandlerAdditionalNikaragva) GetIbftExtraClean(extra *Extra) *Extra {
	return &Extra{
		BlockNumber: extra.BlockNumber,
		Parent:      extra.Parent,
		Validators:  extra.Validators,
		Checkpoint:  extra.Checkpoint,
		Committed:   &Signature{},
		Dummy1:      extra.Dummy1,
		Dummy2:      extra.Dummy2,
	}
}
