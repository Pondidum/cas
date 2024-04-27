package config

func Combine(name string, groups ...*ConfigGroup) *ConfigGroup {

	all := NewConfigGroup(name)

	for _, g := range groups {
		all.flags.AddFlagSet(g.flags)

		for k, v := range g.environment {
			all.environment[k] = v
		}
	}

	return all
}
