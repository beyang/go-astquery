// +build testdata

package service

type ServiceOne struct {
}

func (s *ServiceOne) Get(id string) string {
	DefaultChecker.Check()
	return ""
}

func (s *ServiceOne) List() []string {
	DefaultChecker.Check()
	return nil
}
