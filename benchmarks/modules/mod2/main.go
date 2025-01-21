package main

import "context"

type Mod2 struct {}

func (m *Mod2) Fn(ctx context.Context) (string, error) {
	s := "mod2"
	var depS string
	_ = depS
	var err error
	_ = err

	depS, err = dag.Mod0().Fn(ctx)
	if err != nil {
		return "", err
	}
	s += depS
return s, nil
}
