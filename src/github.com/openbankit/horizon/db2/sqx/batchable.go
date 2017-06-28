package sqx

type Batchable interface {
	// Returns array of params to be inserted/updated
	GetParams() []interface{}
	// Returns hash of the object. Must be immutable
	Hash() uint64
	// Returns true if this and other are equals
	Equals(other interface{}) bool
}

type Updateable interface {
	Batchable
	//Returns key params
	GetKeyParams() []interface{}
}
