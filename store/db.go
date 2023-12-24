package store

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/ttblack/Elastos.ELA.Inscription/constant"
	"math/big"
	"strings"
)

type LevelDBStorage struct {
	db *leveldb.DB
}

func NewLevelDBStorage(dbPath string) (*LevelDBStorage, error) {
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, err
	}

	return &LevelDBStorage{
		db: db,
	}, nil
}

func (s *LevelDBStorage) AddDeployInscription(element *Inscribe) error {
	elements, err := s.GetDeployedInscription()
	if err != nil {
		return err
	}
	if s.IsDuplicateInscribe(element, elements) {
		return errors.New("is duplicate inscribe")
	}
	elements = append(elements, element)
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err = encoder.Encode(elements)
	if err != nil {
		return err
	}
	key := constant.DB_Inscriptions_key
	err = s.db.Put([]byte(key), buffer.Bytes(), nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *LevelDBStorage) IsDuplicateInscribe(element *Inscribe, elements []*Inscribe) bool {
	tick := strings.ToLower(element.Tick)
	for _, ele := range elements {
		if strings.ToLower(ele.Tick) == tick {
			return true
		}
	}
	return false
}

func (s *LevelDBStorage) GetDeployedInscription() ([]*Inscribe, error) {
	key := constant.DB_Inscriptions_key
	data, err := s.get([]byte(key))
	if err != nil && err != leveldb.ErrNotFound {
		return nil, err
	}
	var elements = make([]*Inscribe, 0)
	if err == nil {
		err = gob.NewDecoder(bytes.NewReader(data)).Decode(&elements)
		if err != nil {
			return nil, err
		}
	}
	return elements, nil
}

func (s *LevelDBStorage) GetInscribeInfo(tick string) (*Inscribe, error) {
	elements, err := s.GetDeployedInscription()
	if err != nil {
		return nil, err
	}
	tick = strings.ToLower(tick)
	for _, ele := range elements {
		if strings.ToLower(ele.Tick) == tick {
			return ele, nil
		}
	}
	return nil, errors.New("not found")
}

func (s *LevelDBStorage) get(key []byte) ([]byte, error) {
	value, err := s.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s *LevelDBStorage) AddInscribeTx(height uint64, inscribes []*InscribeTx) error {
	number := big.NewInt(0).SetUint64(height).String()
	key := constant.InScription_PRE + number
	ok, err := s.db.Has([]byte(key), nil)
	if err != nil {
		return err
	}
	if ok {
		return errors.New("allready add " + number)
	}

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err = encoder.Encode(inscribes)
	if err != nil {
		return err
	}
	err = s.db.Put([]byte(key), buffer.Bytes(), nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *LevelDBStorage) GetInscribeTxs(height uint64) ([]*InscribeTx, error) {
	number := big.NewInt(0).SetUint64(height).String()
	key := constant.InScription_PRE + number
	data, err := s.db.Get([]byte(key), nil)
	if err != nil {
		return nil, err
	}
	var elements = make([]*InscribeTx, 0)
	if err == nil {
		err = gob.NewDecoder(bytes.NewReader(data)).Decode(&elements)
		if err != nil {
			return nil, err
		}
	}
	return elements, nil
}

func (s *LevelDBStorage) AddCrossBtcTx(height uint64, inscribes []*InscribeTx) error {
	number := big.NewInt(0).SetUint64(height).String()
	key := constant.InScription_Cross + number
	ok, err := s.db.Has([]byte(key), nil)
	if err != nil {
		return err
	}
	if ok {
		return errors.New("allready add " + number)
	}

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err = encoder.Encode(inscribes)
	if err != nil {
		return err
	}
	err = s.db.Put([]byte(key), buffer.Bytes(), nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *LevelDBStorage) GetInscribeCrossTxs(height uint64) ([]*InscribeTx, error) {
	number := big.NewInt(0).SetUint64(height).String()
	key := constant.InScription_Cross + number
	data, err := s.db.Get([]byte(key), nil)
	if err != nil {
		return nil, err
	}
	var elements = make([]*InscribeTx, 0)
	if err == nil {
		err = gob.NewDecoder(bytes.NewReader(data)).Decode(&elements)
		if err != nil {
			return nil, err
		}
	}
	return elements, nil
}

func (s *LevelDBStorage) Delete(key []byte) error {
	err := s.db.Delete(key, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *LevelDBStorage) Close() error {
	err := s.db.Close()
	if err != nil {
		return err
	}
	return nil
}
