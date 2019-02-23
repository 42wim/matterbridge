package compiler

func (c *Compiler) enterLoop() *Loop {
	loop := &Loop{}

	c.loops = append(c.loops, loop)
	c.loopIndex++

	if c.trace != nil {
		c.printTrace("LOOPE", c.loopIndex)
	}

	return loop
}

func (c *Compiler) leaveLoop() {
	if c.trace != nil {
		c.printTrace("LOOPL", c.loopIndex)
	}

	c.loops = c.loops[:len(c.loops)-1]
	c.loopIndex--
}

func (c *Compiler) currentLoop() *Loop {
	if c.loopIndex >= 0 {
		return c.loops[c.loopIndex]
	}

	return nil
}
