package mediator

import (
	"context"
	"fmt"
	"reflect"
)

type Handler interface{}

type Mediator struct {
	handlers map[reflect.Type]Handler
}

func NewMediator() *Mediator {
	return &Mediator{
		handlers: make(map[reflect.Type]Handler),
	}
}

func (m *Mediator) RegisterHandler(command interface{}, handler Handler) {
	commandType := reflect.TypeOf(command)
	if commandType.Kind() == reflect.Ptr {
		commandType = commandType.Elem()
	}
	m.handlers[commandType] = handler
}

func (m *Mediator) Send(ctx context.Context, command interface{}) (interface{}, error) {
	commandType := reflect.TypeOf(command)
	if commandType.Kind() == reflect.Ptr {
		commandType = commandType.Elem()
	}

	handler, exists := m.handlers[commandType]
	if !exists {
		return nil, fmt.Errorf("no handler registered for command type: %s", commandType.Name())
	}

	handlerValue := reflect.ValueOf(handler)
	handleMethod := handlerValue.MethodByName("Handle")
	if !handleMethod.IsValid() {
		return nil, fmt.Errorf("handler does not implement Handle method")
	}

	args := []reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(command),
	}

	results := handleMethod.Call(args)
	if len(results) != 2 {
		return nil, fmt.Errorf("handle method should return (result, error)")
	}

	if !results[1].IsNil() {
		err, ok := results[1].Interface().(error)
		if !ok {
			return nil, fmt.Errorf("second return value must be an error")
		}
		return nil, err
	}

	return results[0].Interface(), nil
}