package log4go

func (l *Logger) clone() *Logger {
	//只需copy 3个值：flag,level与msg管道，其它沿用默认

	ll := new(Logger)

	if l.formatter != nil {
		ll.formatter = l.formatter
	} else {
		ll.formatter = &JSONFormatter{}
	}
	ll.kv = make(Fields, len(l.kv))
	ll.level = l.level
	ll.flag = l.flag
	ll.handlers = l.handlers

	for k, v := range l.kv {
		ll.kv[k] = v
	}
	return ll
}

var jsonFormatter = new(JSONFormatter)

func (l *Logger) WithField(k string, v interface{}) *Logger {
	ll := l.clone()
	ll.formatter = jsonFormatter
	ll.kv[k] = v
	return ll
}

func (l *Logger) WithFields(kv Fields) *Logger {
	ll := l.clone()
	ll.formatter = jsonFormatter
	for k, v := range kv {
		ll.kv[k] = v
	}
	return ll
}
