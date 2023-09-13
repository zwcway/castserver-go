package database

import (
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/lg"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	log lg.Logger
	db  *gorm.DB
)

func Init(ctx utils.Context, d *gorm.DB) {
	log = ctx.Logger("database")
	// db = d.Debug()
	db = d

	db.AutoMigrate(
		&speaker.Line{},
		&speaker.SpeakerConfig{},
		&speaker.Speaker{},
	)

	speaker.BusGetLines.Register(getLines)
	bus.Register("get line", getLine)
	bus.Register("save line", saveLine).ASync()
	bus.Register("line deleted", deleteLine).ASync()
	bus.Register("line edited", func(o any, a ...any) error {
		line := o.(*speaker.Line)
		um := map[string]any{}
		for i := 0; i < len(a); i += 2 {
			um[a[i].(string)] = a[i+1]
		}
		result := db.Model(line).UpdateColumns(um)
		if result.Error != nil {
			log.Fatal("save line error", lg.Error(result.Error))
			return result.Error
		}
		return nil
	}).ASync()

	bus.Register("get speakers", getSpeakers)
	bus.Register("get speaker", getSpeaker)
	bus.Register("save speaker", saveSpeaker).ASync()
	bus.Register("speaker deleted", deleteSpeaker).ASync()
	speaker.BusSpeakerEdited.Register(func(sp *speaker.Speaker, a ...any) error {
		um := map[string]any{}
		for i := 0; i < len(a); i += 2 {
			um[a[i].(string)] = a[i+1]
		}
		result := db.Model(sp).UpdateColumns(um)
		if result.Error != nil {
			log.Fatal("save speaker error", lg.Error(result.Error))
			return result.Error
		}
		return nil
	}).ASync()
}

func getLines(lineList *[]*speaker.Line) error {
	lines := []speaker.Line{}
	result := db.Find(&lines)
	if result.RowsAffected > 0 {
		for i := 0; i < len(lines); i++ {
			*lineList = append(*lineList, &lines[i])
		}
		return nil
	}
	if result.Error != nil {
		log.Fatal("read all lines error", lg.Error(result.Error))
	}
	return result.Error
}

func getLine(o any, a ...any) error {
	line := a[0].(*speaker.Line)
	result := db.Take(line)
	if result.Error != nil {
		log.Fatal("read line error", lg.Uint("line", uint64(line.ID)), lg.Error(result.Error))
	}
	return result.Error
}

func saveLine(o any, a ...any) error {
	line := o.(*speaker.Line)
	result := db.Save(line)

	if result.Error != nil {
		log.Fatal("save line error", lg.Uint("line", uint64(line.ID)), lg.Error(result.Error))
	}
	return result.Error
}

func deleteLine(o any, a ...any) error {
	line := o.(*speaker.Line)
	result := db.Delete(line)
	if result.Error != nil {
		log.Fatal("delete line error", lg.Uint("line", uint64(line.ID)), lg.Error(result.Error))
	}
	return result.Error
}

func getSpeakers(o any, a ...any) error {
	spList := a[0].(*[]*speaker.Speaker)
	sps := []speaker.Speaker{}
	result := db.Preload(clause.Associations).Find(&sps)
	if result.RowsAffected > 0 {
		for i := 0; i < len(sps); i++ {
			*spList = append(*spList, &sps[i])
		}
		return nil
	}
	if result.Error != nil {
		log.Fatal("read all lines error", lg.Error(result.Error))
	}
	return result.Error
}

func getSpeaker(o any, a ...any) error {
	sp := a[0].(*speaker.Speaker)
	result := db.Take(sp)
	if result.Error != nil {
		log.Fatal("read speaker error", lg.Uint("speaker", uint64(sp.ID)), lg.Error(result.Error))
	}
	return result.Error
}

func saveSpeaker(o any, a ...any) error {
	sp := a[0].(*speaker.Speaker)
	result := db.Session(&gorm.Session{FullSaveAssociations: true}).Save(sp)
	if result.Error != nil {
		log.Fatal("save speaker error", lg.Uint("speaker", uint64(sp.ID)), lg.Error(result.Error))
	}
	return result.Error
}

func deleteSpeaker(o any, a ...any) error {
	sp := a[0].(*speaker.Speaker)
	result := db.Delete(sp)
	if result.Error != nil {
		log.Fatal("delete speaker error", lg.Uint("speaker", uint64(sp.ID)), lg.Error(result.Error))
	}
	return result.Error
}
