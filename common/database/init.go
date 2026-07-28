package database

import (
	lg "github.com/zwcway/castserver-go/common/log"
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
	speaker.BusGetLine.Register(getLine)
	speaker.BusLineSaved.Register(saveLine).ASync()
	speaker.BusLineDeleted.Register(deleteLine).ASync()
	speaker.BusLineEdited.Register(func(line *speaker.Line, field string, value, _ any) error {
		um := map[string]any{field: value}
		result := db.Model(line).UpdateColumns(um)
		if result.Error != nil {
			log.Fatal("save line error", lg.Error(result.Error))
			return result.Error
		}
		return nil
	}).ASync()

	speaker.BusGetSpeakers.Register(getSpeakers)
	speaker.BusGetSpeaker.Register(getSpeaker)
	speaker.BusSpeakerSaved.Register(saveSpeaker)
	speaker.BusSpeakerDeleted.Register(deleteSpeaker).ASync()
	speaker.BusSpeakerEdited.Register(func(sp *speaker.Speaker, field string, value, _ any) error {
		um := map[string]any{field: value}
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

func getLine(line *speaker.Line) error {
	result := db.Take(line)
	if result.Error != nil {
		log.Fatal("read line error", lg.Uint("line", uint64(line.ID)), lg.Error(result.Error))
	}
	return result.Error
}

func saveLine(line *speaker.Line) error {
	result := db.Save(line)

	if result.Error != nil {
		log.Fatal("save line error", lg.Uint("line", uint64(line.ID)), lg.Error(result.Error))
	}
	return result.Error
}

func deleteLine(line, _ *speaker.Line) error {
	result := db.Delete(line)
	if result.Error != nil {
		log.Fatal("delete line error", lg.Uint("line", uint64(line.ID)), lg.Error(result.Error))
	}
	return result.Error
}

func getSpeakers(spList *[]*speaker.Speaker) error {
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

func getSpeaker(sp *speaker.Speaker) error {
	result := db.Take(sp)
	if result.Error != nil {
		log.Fatal("read speaker error", lg.Uint("speaker", uint64(sp.ID)), lg.Error(result.Error))
	}
	return result.Error
}

func saveSpeaker(sp *speaker.Speaker) error {
	result := db.Session(&gorm.Session{FullSaveAssociations: true}).Save(sp)
	if result.Error != nil {
		log.Fatal("save speaker error", lg.Uint("speaker", uint64(sp.ID)), lg.Error(result.Error))
	}
	return result.Error
}

func deleteSpeaker(sp *speaker.Speaker) error {
	result := db.Delete(sp)
	if result.Error != nil {
		log.Fatal("delete speaker error", lg.Uint("speaker", uint64(sp.ID)), lg.Error(result.Error))
	}
	return result.Error
}
