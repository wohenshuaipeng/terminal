package common

type Emitter interface {
    Emit(event string, payload any)
}

type NopEmitter struct{}

func (NopEmitter) Emit(string, any) {}
