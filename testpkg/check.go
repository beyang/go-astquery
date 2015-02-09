package service

var DefaultChecker = Checker{}

type Checker struct{}

// Checker.check should be called in each service method
func (c *Checker) Check() {}
