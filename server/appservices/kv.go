package appservices

func (a *AppServices) KVSet(botUserID, prefix, id string, ref interface{}) (bool, error) {
	if err := a.ensureFromBot(botUserID); err != nil {
		return false, err
	}
	return a.store.AppKV.Set(botUserID, prefix, id, ref)
}

func (a *AppServices) KVGet(botUserID, prefix, id string, ref interface{}) error {
	if err := a.ensureFromBot(botUserID); err != nil {
		return err
	}
	return a.store.AppKV.Get(botUserID, prefix, id, ref)
}

func (a *AppServices) KVDelete(botUserID, prefix, id string) error {
	if err := a.ensureFromBot(botUserID); err != nil {
		return err
	}
	return a.store.AppKV.Delete(botUserID, prefix, id)
}

func (a *AppServices) KVList(botUserID, prefix string, processf func(key string) error) error {
	if err := a.ensureFromBot(botUserID); err != nil {
		return err
	}
	return a.store.AppKV.List(botUserID, prefix, processf)
}
