package register

type rNode struct {
	rType  string
	rValue string
}

type stack struct {
	node []rNode
}

func newRegisterStack() *stack {
	return &stack{node: []rNode{}}
}

func (r *stack) push(rType, rValue string) {
	r.node = append(r.node, rNode{
		rType:  rType,
		rValue: rValue,
	})
}

func (r *stack) pop() (string, string) {
	if len(r.node) <= 0 {
		return "", ""
	}
	rNode := r.node[len(r.node)-1]
	r.node = r.node[:len(r.node)-1]
	return rNode.rType, rNode.rValue
}

func (r *stack) peek() (string, string) {
	if len(r.node) <= 0 {
		return "", ""
	}

	rNode := r.node[len(r.node)-1]
	return rNode.rType, rNode.rValue
}

func (r *stack) len() int {
	return len(r.node)
}
