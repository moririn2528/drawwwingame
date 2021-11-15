package domain

// type MapList struct {
// 	key   []interface{}
// 	value []interface{}
// }

// func NewMapList() *MapList {
// 	return new(MapList)
// }
// func NewMapListInitMap(input_map map[interface{}]interface{}) *MapList {
// 	m := new(MapList)
// 	for k, v := range input_map {
// 		m.Append(k, v)
// 	}
// 	return m
// }

// func (m *MapList) Append(key interface{}, value interface{}) {
// 	m.key = append(m.key, key)
// 	m.value = append(m.value, value)
// }

// func (m *MapList) Clear() {
// 	m.key = make([]interface{}, 0, 0)
// 	m.value = make([]interface{}, 0, 0)
// }

// func (m *MapList) Merge(ml_list ...*MapList) *MapList {
// 	for _, a := range ml_list {
// 		m.key = append(m.key, a.key)
// 		m.value = append(m.value, a.value)
// 		a.Clear()
// 	}
// 	return m
// }

// func (m *MapList) ToMap() map[interface{}]interface{} {
// 	s := make(map[interface{}]interface{})
// 	for i, k := range m.key {
// 		s[k] = m.value[i]
// 	}
// 	return s
// }
