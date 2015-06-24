package main

import "time"

type MailDatabase []MailConnection

func (db *MailDatabase) cleanupDatabase() {
	//	mc.mu.Lock()
	new := MailDatabase{}
	for _, v := range *db {
		if time.Since(v.expireStamp).Seconds() < 0 {
			new = append(new, v)
		}
	}
	db = &new
	//	mc.mu.Unlock()
}

func (db *MailDatabase) save(mc MailConnection) {
	//config.mu.Lock()
	new := append(*db, mc)
	db = &new
	//config.mu.Unlock()

}
