package id

type AccountId string

func (a AccountId) String() string {
	return string(a)
}

type TripId string

func (t TripId) String() string {
	return string(t)
}

type IdentityId string

func (i IdentityId) String() string {
	return string(i)
}

type CarId string

func (c CarId) String() string {
	return string(c)
}