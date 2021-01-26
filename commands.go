package bcr

import "sort"

// Commands is a sortable slice of Command
type Commands []*Command

func (c Commands) Len() int      { return len(c) }
func (c Commands) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c Commands) Less(i, j int) bool {
	return sort.StringsAreSorted([]string{c[i].Name, c[j].Name})
}
