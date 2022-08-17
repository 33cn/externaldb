package proofconfig

import (
	"encoding/json"

	lru "github.com/hashicorp/golang-lru"

	"github.com/33cn/externaldb/db"
)

var OpenAccessControl = false
var memberCache *lru.Cache
var orgCache *lru.Cache

func init() {
	var err error
	memberCache, err = lru.New(1024)
	if err != nil {
		panic(err)
	}
	orgCache, err = lru.New(1024)
	if err != nil {
		panic(err)
	}
}

// ConfigDB ConfigDB
type configDB struct {
	dbX, tableX       string
	orgDBX, orgTableX string
	delDBX, delTableX string
	db                db.WrapDB
}

func InitOpenAccessControl(isOpen bool) {
	OpenAccessControl = isOpen
}

// NewConfigDB NewConfigDB
func NewConfigDB(db db.WrapDB) ConfigDB {
	d := &configDB{
		db:        db,
		dbX:       DBX,
		tableX:    TableX,
		delDBX:    DeleteDBX,
		delTableX: DeleteTableX,
		orgDBX:    OrgDBX,
		orgTableX: OrgTableX,
	}
	if OpenAccessControl {
		return d
	}
	return &None{d}
}

func (db1 *configDB) GetMember(addr string) (*Member, error) {
	if m, ok := memberCache.Get(addr); ok {
		return m.(*Member), nil
	}
	m1, err := db1.db.Get(db1.dbX, db1.tableX, MemberID(addr))
	if err != nil {
		return nil, err
	}
	var mem Member
	err = json.Unmarshal(*m1, &mem)
	if err != nil {
		return nil, err
	}
	memberCache.Add(addr, &mem)
	return &mem, nil
}

func (db1 *configDB) GetMemberDel(addr string, h, i int64) (*memberDel, error) {
	m1, err := db1.db.Get(db1.delDBX, db1.delTableX, memberDelID(addr, h, i))
	if err != nil {
		return nil, err
	}
	var mem memberDel
	err = json.Unmarshal(*m1, &mem)
	if err != nil {
		return nil, err
	}
	return &mem, nil
}

func (db1 *configDB) GetOrganization(org string) (*Organization, error) {
	if o, ok := orgCache.Get(org); ok {
		return o.(*Organization), nil
	}
	m1, err := db1.db.Get(db1.orgDBX, db1.orgTableX, OrganizationID(org))
	if err != nil {
		return nil, err
	}
	var o Organization
	err = json.Unmarshal(*m1, &o)
	if err != nil {
		return nil, err
	}
	orgCache.Add(org, &o)
	return &o, nil
}

func (db1 *configDB) SetMember(addr string, m *dbMember) error {
	memberCache.Add(addr, m)
	return db1.db.Set(db1.dbX, db1.tableX, m.M.ID(), m)
}

func (db1 *configDB) SetMemberDel(addr string, m *dbMemberDel) error {
	return db1.db.Set(db1.delDBX, db1.delTableX, m.M.ID(), m)
}

func (db1 *configDB) SetOrganization(org string, o *dbOrganization) error {
	memberCache.Add(org, o)
	return db1.db.Set(db1.orgDBX, db1.orgTableX, o.M.ID(), o)
}

// IsHaveProofPermission check Permission
func (db1 *configDB) IsHaveProofPermission(addr string) bool {
	_, err := db1.GetMember(addr)
	if err != nil {
		log.Error("get member failed", "err", err, "address", addr)
		return false
	}
	return true
}

// GetOrganizationName Get Organization by member
func (db1 *configDB) GetOrganizationName(addr string) (string, error) {
	m, err := db1.GetMember(addr)
	if err != nil {
		log.Error("get member failed", "err", err, "address", addr)
		return "", err
	}
	return m.Organization, nil
}

// IsHaveDelProofPermission check  DelProofPermission
func (db1 *configDB) IsHaveDelProofPermission(send, proofOrg, proofOwner string) bool {
	m, err := db1.GetMember(send)
	if err != nil {
		log.Error("get member failed", "err", err, "address", send)
		return false
	}
	return m.PrivilegeToProof(proofOrg, proofOwner)
}
