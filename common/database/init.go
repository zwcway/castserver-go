package database

import (
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	log *zap.Logger
	db  *gorm.DB
)

func Init(ctx utils.Context, d *gorm.DB) {
	log = ctx.Logger("database")
	db = d

	db.AutoMigrate(&speaker.Line{}, &speaker.Speaker{}, &speaker.SpeakerConfig{})

	bus.Register("get lines", getLines)
	bus.Register("get line", getLine)
	bus.Register("save line", saveLine).ASync()
	bus.Register("line deleted", deleteLine).ASync()
	bus.Register("line edited", func(a ...any) error {
		line := a[0].(*speaker.Line)
		um := map[string]any{}
		for i := 1; i < len(a); i += 2 {
			um[a[i].(string)] = a[i+1]
		}
		result := db.Model(line).UpdateColumns(um)
		if result.Error != nil {
			log.Fatal("save line error", zap.Error(result.Error))
			return result.Error
		}
		return nil
	}).ASync()

	bus.Register("get speakers", getSpeakers)
	bus.Register("get speaker", getSpeaker)
	bus.Register("save speaker", saveSpeaker).ASync()
	bus.Register("speaker deleted", deleteSpeaker).ASync()
	bus.Register("speaker edited", func(a ...any) error {
		sp := a[0].(*speaker.Speaker)
		um := map[string]any{}
		for i := 1; i < len(a); i += 2 {
			um[a[i].(string)] = a[i+1]
		}
		result := db.Model(sp).UpdateColumns(um)
		if result.Error != nil {
			log.Fatal("save speaker error", zap.Error(result.Error))
			return result.Error
		}
		return nil
	}).ASync()
}

func getLines(a ...any) error {
	lineList := a[0].(*[]*speaker.Line)
	lines := []speaker.Line{}
	result := db.Find(&lines)
	if result.RowsAffected > 0 {
		for i := 0; i < len(lines); i++ {
			*lineList = append(*lineList, &lines[i])
		}
		return nil
	}
	if result.Error != nil {
		log.Fatal("read all lines error", zap.Error(result.Error))
	}
	return result.Error
}

func getLine(a ...any) error {
	line := a[0].(*speaker.Line)
	result := db.Take(line)
	if result.Error != nil {
		log.Fatal("read line error", zap.Uint8("line", line.ID), zap.Error(result.Error))
	}
	return result.Error
}

func saveLine(a ...any) error {
	line := a[0].(*speaker.Line)
	result := db.Save(line)

	if result.Error != nil {
		log.Fatal("save line error", zap.Uint8("line", line.ID), zap.Error(result.Error))
	}
	return result.Error
}

func deleteLine(a ...any) error {
	line := a[0].(*speaker.Line)
	result := db.Delete(line)
	if result.Error != nil {
		log.Fatal("delete line error", zap.Uint8("line", line.ID), zap.Error(result.Error))
	}
	return result.Error
}

func getSpeakers(a ...any) error {
	spList := a[0].(*[]*speaker.Speaker)
	sps := []speaker.Speaker{}
	result := db.Find(&sps)
	if result.RowsAffected > 0 {
		for i := 0; i < len(sps); i++ {
			*spList = append(*spList, &sps[i])
		}
		return nil
	}
	if result.Error != nil {
		log.Fatal("read all lines error", zap.Error(result.Error))
	}
	return result.Error
}

func getSpeaker(a ...any) error {
	sp := a[0].(*speaker.Speaker)
	result := db.Take(sp)
	if result.Error != nil {
		log.Fatal("read speaker error", zap.Uint32("speaker", sp.ID), zap.Error(result.Error))
	}
	return result.Error
}

func saveSpeaker(a ...any) error {
	sp := a[0].(*speaker.Speaker)
	result := db.Save(sp)
	if result.Error != nil {
		log.Fatal("save speaker error", zap.Uint32("speaker", sp.ID), zap.Error(result.Error))
	}
	return result.Error
}

func deleteSpeaker(a ...any) error {
	sp := a[0].(*speaker.Speaker)
	result := db.Delete(sp)
	if result.Error != nil {
		log.Fatal("delete speaker error", zap.Uint32("speaker", sp.ID), zap.Error(result.Error))
	}
	return result.Error
}
