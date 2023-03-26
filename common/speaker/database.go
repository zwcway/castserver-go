package speaker

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/dsp"
	"github.com/zwcway/castserver-go/common/jsonpack"
)

type DBeqData struct {
	Eq *dsp.DataProcess
}

func (j *DBeqData) GormDataType() string {
	return "blob"
}

// 实现 sql.Scanner 接口，允许出库
func (j *DBeqData) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal value", value))
	}

	result := dsp.DataProcess{}
	err := jsonpack.Unmarshal(bytes, &result)
	if err != nil {
		j.Eq = dsp.NewDataProcess(0)
		return err
	}
	j.Eq = &result
	return nil
}

// 实现 driver.Valuer 接口，允许入库
func (j DBeqData) Value() (driver.Value, error) {
	if j.Eq == nil {
		return nil, nil
	}
	return jsonpack.Marshal(j.Eq)
}

type DBChannelRoute struct {
	R []audio.ChannelRoute
}

func (j *DBChannelRoute) GormDataType() string {
	return "blob"
}

// 实现 sql.Scanner 接口，允许出库
func (j *DBChannelRoute) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal value", value))
	}

	result := []audio.ChannelRoute{}
	err := jsonpack.Unmarshal(bytes, &result)
	if err != nil {
		j.R = []audio.ChannelRoute{}
		return err
	}
	j.R = result
	return nil
}

// 实现 driver.Valuer 接口，允许入库
func (j DBChannelRoute) Value() (driver.Value, error) {
	if len(j.R) == 0 {
		return nil, nil
	}
	return jsonpack.Marshal(j)
}
