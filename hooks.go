package gorm

type BeforeCreateInterface interface {
	BeforeCreate(*DB) error
}

type AfterCreateInterface interface {
	AfterCreate(*DB) error
}

type BeforeUpdateInterface interface {
	BeforeUpdate(*DB) error
}

type AfterUpdateInterface interface {
	AfterUpdate(*DB) error
}

type BeforeQueryInterface interface {
	BeforeQuery(*DB) error
}

type AfterQueryInterface interface {
	AfterQuery(*DB) error
}

type BeforeDeleteInterface interface {
	BeforeDelete(*DB) error
}

type AfterDeleteInterface interface {
	AfterDelete(*DB) error
}

type hooks struct {
	bc BeforeCreateInterface
	ac AfterCreateInterface
	bu BeforeUpdateInterface
	au AfterUpdateInterface
	bq BeforeQueryInterface
	aq AfterQueryInterface
	bd BeforeDeleteInterface
	ad AfterDeleteInterface
}

func (db *DB) parseHooks(obj interface{}) *DB {
	if db.hks != nil {
		return db
	}
	hks := new(hooks)
	if i, ok := obj.(BeforeCreateInterface); ok {
		hks.bc = i
	}

	if i, ok := obj.(AfterCreateInterface); ok {
		hks.ac = i
	}
	if i, ok := obj.(BeforeUpdateInterface); ok {
		hks.bu = i
	}

	if i, ok := obj.(AfterUpdateInterface); ok {
		hks.au = i
	}
	if i, ok := obj.(BeforeQueryInterface); ok {
		hks.bq = i
	}

	if i, ok := obj.(AfterQueryInterface); ok {
		hks.aq = i
	}
	if i, ok := obj.(BeforeDeleteInterface); ok {
		hks.bd = i
	}

	if i, ok := obj.(AfterDeleteInterface); ok {
		hks.ad = i
	}
	db.hks = hks
	return db
}

func (hks *hooks) SetHooksOnCreate() bool {
	return hks.bc != nil || hks.ac != nil
}

func (hks *hooks) SetHooksOnUpdate() bool {
	return hks.bu != nil || hks.au != nil
}

func (hks *hooks) SetHooksOnQuery() bool {
	return hks.bq != nil || hks.aq != nil
}

func (hks *hooks) SetHooksOnDelete() bool {
	return hks.bd != nil || hks.ad != nil
}

func (hks *hooks) GetBeforeCreateHook() BeforeCreateInterface {
	return hks.bc
}

func (hks *hooks) GetAfterCreateHook() AfterCreateInterface {
	return hks.ac
}

func (hks *hooks) GetBeforeUpdateHook() BeforeUpdateInterface {
	return hks.bu
}

func (hks *hooks) GetAfterUpdateHook() AfterUpdateInterface {
	return hks.au
}
func (hks *hooks) GetBeforeQueryHook() BeforeQueryInterface {
	return hks.bq
}

func (hks *hooks) GetAfterQueryHook() AfterQueryInterface {
	return hks.aq
}
func (hks *hooks) GetBeforeDeleteHook() BeforeDeleteInterface {
	return hks.bd
}

func (hks *hooks) GetAfterDeleteHook() AfterDeleteInterface {
	return hks.ad
}
