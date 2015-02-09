// +build testdata

package service

type ServiceTwo struct {
}

func (s *ServiceTwo) Get(id string) string {
	DefaultChecker.Check()
	return ""
}

func (s *ServiceTwo) List() []string {
	DefaultChecker.Check()
	return nil
}

func (s *ServiceTwo) UncheckedMeth() {}
