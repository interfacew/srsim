package huohuo

import (
	"github.com/simimpact/srsim/pkg/engine"
	"github.com/simimpact/srsim/pkg/engine/info"
	"github.com/simimpact/srsim/pkg/engine/target/character"
	"github.com/simimpact/srsim/pkg/key"
	"github.com/simimpact/srsim/pkg/model"
)

func init() {
	character.Register(key.Huohuo, character.Config{
		Create:     NewInstance,
		Rarity:     5,
		Element:    model.DamageType_WIND,
		Path:       model.Path_ABUNDANCE,
		MaxEnergy:  140,
		Promotions: promotions,
		Traces:     traces,
		SkillInfo: character.SkillInfo{
			Attack: character.Attack{
				SPAdd:      1,
				TargetType: model.TargetType_ENEMIES,
			},
			Skill: character.Skill{
				SPNeed:     1,
				TargetType: model.TargetType_ALLIES,
			},
			Ult: character.Ult{
				TargetType: model.TargetType_ALLIES,
			},
			Technique: character.Technique{
				TargetType: model.TargetType_ENEMIES,
				IsAttack:   false,
			},
		},
	})
}

type char struct {
	engine      engine.Engine
	id          key.TargetID
	info        info.Character
	DispelCount int
	TalentRound int
}

func NewInstance(engine engine.Engine, id key.TargetID, charInfo info.Character) info.CharInstance {
	c := &char{
		engine:      engine,
		id:          id,
		info:        charInfo,
		DispelCount: 0,
		TalentRound: 0,
	}
	engine.Events().ActionStart.Subscribe(c.TalentActionStartListener)
	return c
}
