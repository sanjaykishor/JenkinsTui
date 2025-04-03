package utils

import "github.com/sanjaykishor/JenkinsTui.git/internal/api"


// countFreeNodes returns the number of online and idle nodes
func CountFreeNodes(nodes []api.Node) int {
	count := 0
	for _, node := range nodes {
		if node.Online && node.Idle {
			count++
		}
	}
	return count
}