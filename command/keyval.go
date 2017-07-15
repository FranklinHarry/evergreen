package command

import (
	"context"
	"fmt"

	"github.com/evergreen-ci/evergreen/model"
	"github.com/evergreen-ci/evergreen/rest/client"
	"github.com/evergreen-ci/evergreen/util"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

type keyValInc struct {
	Key         string `mapstructure:"key"`
	Destination string `mapstructure:"destination"`
}

func keyValIncFactory() Command     { return &keyValInc{} }
func (c *keyValInc) Name() string   { return "inc" }
func (c *keyValInc) Plugin() string { return "keyval" }

// ParseParams validates the input to the keyValInc, returning an error
// if something is incorrect. Fulfills Command interface.
func (c *keyValInc) ParseParams(params map[string]interface{}) error {
	err := mapstructure.Decode(params, c)
	if err != nil {
		return err
	}

	if c.Key == "" || c.Destination == "" {
		return fmt.Errorf("error parsing '%v' params: key and destination may not be blank",
			c.Name())
	}

	return nil
}

// Execute fetches the expansions from the API server
func (c *keyValInc) Execute(ctx context.Context,
	comm client.Communicator, logger client.LoggerProducer, conf *model.TaskConfig) error {

	if err := util.ExpandValues(c, conf.Expansions); err != nil {
		return err
	}

	td := client.TaskData{ID: conf.Task.Id, Secret: conf.Task.Secret}
	keyVal := model.KeyVal{Key: c.Key}
	err := comm.IncrementKey(ctx, td, &keyVal) //.TaskPostJSON(IncRoute, c.Key)
	if err != nil {
		return errors.Wrapf(err, "problem incriminating key %s", c.Key)
	}

	conf.Expansions.Put(c.Destination, fmt.Sprintf("%d", keyVal.Value))
	return nil
}