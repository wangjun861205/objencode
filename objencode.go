package objencode

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"reflect"
)

func Encode(i interface{}) ([]byte, error) {
	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	val := reflect.ValueOf(i)
	switch val.Kind() {
	case reflect.Ptr:
		bContent, err := Encode(val.Elem().Interface())
		if err != nil {
			return bContent, err
		}
		buffer.Write(bContent)
		return buffer.Bytes(), nil
	case reflect.Struct:
		subBuffer := bytes.NewBuffer(make([]byte, 0, 1024))
		nField := val.NumField()
		for i := 0; i < nField; i++ {
			field := val.Field(i)
			bContent, err := Encode(field.Interface())
			if err != nil {
				return bContent, err
			}
			subBuffer.Write(bContent)
		}
		err := writeLength(subBuffer, buffer)
		if err != nil {
			return []byte{}, err
		}
		buffer.Write(subBuffer.Bytes())
		return buffer.Bytes(), nil
	case reflect.Array, reflect.Slice:
		subBuffer := bytes.NewBuffer(make([]byte, 0, 1024))
		l := val.Len()
		for i := 0; i < l; i++ {
			bContent, err := Encode(val.Index(i).Interface())
			if err != nil {
				return bContent, err
			}
			subBuffer.Write(bContent)
		}
		err := writeLength(subBuffer, buffer)
		if err != nil {
			return []byte{}, err
		}
		buffer.Write(subBuffer.Bytes())
		return buffer.Bytes(), nil
	case reflect.String:
		s := val.Interface().(string)
		err := writeLength(bytes.NewBuffer([]byte(s)), buffer)
		if err != nil {
			return []byte{}, err
		}
		buffer.Write([]byte(s))
		return buffer.Bytes(), nil
	case reflect.Map:
		subBuffer := bytes.NewBuffer(make([]byte, 0, 1024))
		keys := val.MapKeys()
		for _, key := range keys {
			keyContent, err := Encode(key.Interface())
			if err != nil {
				return keyContent, err
			}
			subBuffer.Write(keyContent)
			valContent, err := Encode(val.MapIndex(key).Interface())
			if err != nil {
				return valContent, err
			}
			subBuffer.Write(valContent)
		}
		err := writeLength(subBuffer, buffer)
		if err != nil {
			return []byte{}, err
		}
		buffer.Write(subBuffer.Bytes())
		return buffer.Bytes(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var i64 int64
		switch i := val.Interface().(type) {
		case int:
			i64 = int64(i)
		case int8:
			i64 = int64(i)
		case int16:
			i64 = int64(i)
		case int32:
			i64 = int64(i)
		case int64:
			i64 = i
		default:
			return []byte{}, errors.New("Unknow errror")
		}
		subBuffer := bytes.NewBuffer(make([]byte, 0, 8))
		err := binary.Write(subBuffer, binary.LittleEndian, i64)
		if err != nil {
			return []byte{}, err
		}
		err = writeLength(subBuffer, buffer)
		if err != nil {
			return []byte{}, err
		}
		buffer.Write(subBuffer.Bytes())
		return buffer.Bytes(), nil
	case reflect.Float32, reflect.Float64:
		var f64 float64
		switch f := val.Interface().(type) {
		case float32:
			f64 = float64(f)
		case float64:
			f64 = f
		default:
			return []byte{}, errors.New("Unknow error")
		}
		subBuffer := bytes.NewBuffer(make([]byte, 0, 8))
		binary.Write(subBuffer, binary.LittleEndian, f64)
		err := writeLength(subBuffer, buffer)
		if err != nil {
			return []byte{}, err
		}
		buffer.Write(subBuffer.Bytes())
		return buffer.Bytes(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var ui64 uint64
		switch ui := val.Interface().(type) {
		case uint:
			ui64 = uint64(ui)
		case uint8:
			ui64 = uint64(ui)
		case uint16:
			ui64 = uint64(ui)
		case uint32:
			ui64 = uint64(ui)
		case uint64:
			ui64 = ui
		}
		subBuffer := bytes.NewBuffer(make([]byte, 0, 8))
		err := binary.Write(subBuffer, binary.LittleEndian, ui64)
		if err != nil {
			return []byte{}, err
		}
		err = writeLength(subBuffer, buffer)
		if err != nil {
			return []byte{}, err
		}
		buffer.Write(subBuffer.Bytes())
		return buffer.Bytes(), nil
	case reflect.Complex64, reflect.Complex128:
		var c128 complex128
		switch c := val.Interface().(type) {
		case complex64:
			c128 = complex128(c)
		case complex128:
			c128 = c
		default:
			return []byte{}, errors.New("Unknow error")
		}
		subBuffer := bytes.NewBuffer(make([]byte, 0, 16))
		binary.Write(subBuffer, binary.LittleEndian, c128)
		err := writeLength(subBuffer, buffer)
		if err != nil {
			return []byte{}, err
		}
		buffer.Write(subBuffer.Bytes())
		return buffer.Bytes(), nil
	case reflect.Bool:
		b := val.Interface().(bool)
		subBuffer := bytes.NewBuffer(make([]byte, 0, 1))
		err := binary.Write(subBuffer, binary.LittleEndian, b)
		if err != nil {
			return []byte{}, err
		}
		err = writeLength(subBuffer, buffer)
		if err != nil {
			return []byte{}, err
		}
		buffer.Write(subBuffer.Bytes())
		return buffer.Bytes(), nil
	default:
		return buffer.Bytes(), errors.New("Not supported type")
	}
}

func Decode(b []byte, obj interface{}) error {
	if reflect.ValueOf(obj).Kind() != reflect.Ptr {
		return errors.New("Should pass a object pointer as argument")
	}
	val := reflect.ValueOf(obj).Elem()
	switch val.Kind() {
	case reflect.Ptr:
		err := Decode(b, val.Interface())
		if err != nil {
			return err
		}
	case reflect.Struct:
		_, content, _, err := readContent(b)
		if err != nil {
			return err
		}
		nField := val.NumField()
		for i := 0; i < nField; i++ {
			_, fieldContent, fieldRemain, err := readContent(content)
			if err != nil {
				return err
			}
			err = Decode(fieldContent, val.Field(i).Addr().Interface())
			if err != nil {
				return err
			}
			content = fieldRemain
		}
		return nil
	case reflect.Array:
		elemType := val.Type().Elem()
		elemNum := val.Len()
		arrayType := reflect.ArrayOf(int(elemNum), elemType)
		array := reflect.New(arrayType)
		for i := 0; i < elemNum; i++ {
			_, elemContent, elemRemain, err := readContent(b)
			if err != nil {
				return err
			}
			err = Decode(elemContent, array.Elem().Index(i).Addr().Interface())
			if err != nil {
				return err
			}
			b = elemRemain
		}
		val.Set(array.Elem())
		return nil
	case reflect.Slice:
		elemType := val.Type().Elem()
		sliceType := reflect.SliceOf(elemType)
		slice := reflect.MakeSlice(sliceType, 0, 32)
		for {
			_, elemContent, elemRemain, err := readContent(b)
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			elem := reflect.New(elemType)
			err = Decode(elemContent, elem.Interface())
			if err != nil {
				return err
			}
			slice = reflect.Append(slice, elem.Elem())
			if len(elemRemain) == 0 {
				break
			}
			b = elemRemain
		}
		val.Set(slice)
		return nil
	case reflect.Map:
		keyType := val.Type().Key()
		valType := val.Type().Elem()
		mapType := reflect.MapOf(keyType, valType)
		m := reflect.MakeMap(mapType)
		for {
			_, keyContent, remain, err := readContent(b)
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			newKey := reflect.New(keyType)
			err = Decode(keyContent, newKey.Interface())
			if err != nil {
				return err
			}
			_, valContent, remain, err := readContent(remain)
			if err != nil && err != io.EOF {
				return err
			}
			newVal := reflect.New(valType)
			err = Decode(valContent, newVal.Interface())
			if err != nil {
				return err
			}
			m.SetMapIndex(newKey.Elem(), newVal.Elem())
			if len(remain) == 0 {
				break
			}
			b = remain
		}
		val.Set(m)
		return nil
	case reflect.String:
		val.SetString(string(b[:]))
		return nil
	case reflect.Bool:
		var bVal bool
		err := binary.Read(bytes.NewReader(b), binary.LittleEndian, &bVal)
		if err != nil {
			return err
		}
		val.SetBool(bVal)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var i64 int64
		err := binary.Read(bytes.NewReader(b), binary.LittleEndian, &i64)
		if err != nil {
			return err
		}
		val.SetInt(i64)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var u64 uint64
		err := binary.Read(bytes.NewReader(b), binary.LittleEndian, &u64)
		if err != nil {
			return err
		}
		val.SetUint(u64)
		return nil
	case reflect.Float32, reflect.Float64:
		var f64 float64
		err := binary.Read(bytes.NewReader(b), binary.LittleEndian, &f64)
		if err != nil {
			return err
		}
		val.SetFloat(f64)
		return nil
	case reflect.Complex64, reflect.Complex128:
		var c128 complex128
		err := binary.Read(bytes.NewReader(b), binary.LittleEndian, &c128)
		if err != nil {
			return err
		}
		val.SetComplex(c128)
		return nil
	default:
		return errors.New("Not supported type")
	}
	return nil
}

func getLength(b []byte) (int, error) {
	lenByte := bytes.NewBuffer(b[:8])
	var length int64
	err := binary.Read(lenByte, binary.LittleEndian, &length)
	if err != nil {
		return 0, err
	}
	return int(length), nil
}

func readContent(b []byte) (int, []byte, []byte, error) {
	reader := bytes.NewReader(b)
	lenByte := make([]byte, 8)
	_, err := reader.Read(lenByte)
	if err != nil {
		return 0, []byte{}, []byte{}, err
	}
	var l int64
	err = binary.Read(bytes.NewReader(lenByte), binary.LittleEndian, &l)
	if err != nil {
		return 0, []byte{}, []byte{}, err
	}
	length := int(l)
	content := make([]byte, length)
	_, err = reader.Read(content)
	if err != nil {
		return 0, []byte{}, []byte{}, err
	}
	remain, err := ioutil.ReadAll(reader)
	if err != nil {
		return 0, []byte{}, []byte{}, err
	}
	return length, content, remain, nil
}

func writeLength(rdBuf, wtBuf *bytes.Buffer) error {
	length := int64(rdBuf.Len())
	lenBuf := bytes.NewBuffer(make([]byte, 0))
	err := binary.Write(lenBuf, binary.LittleEndian, length)
	if err != nil {
		return err
	}
	_, err = wtBuf.Write(lenBuf.Bytes())
	if err != nil {
		return err
	}
	return nil
}
