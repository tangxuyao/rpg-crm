package crm

import (
	"golang.org/x/net/context"
	"proto/crm"
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"strconv"
	"gopkg.in/mgo.v2"
	"../../MongoData"
	"gopkg.in/mgo.v2/bson"
	"time"
	"log"
)

type CRMService struct {
	pIndex int64
	kv     *consul.KV
	mgo    *mgo.Session
}

func (s *CRMService) Init(consulURL string, mgoUrl string) {
	fmt.Println("[CRM] 初始化系统...")
	config := consul.DefaultConfig()
	config.Address = consulURL
	c, err := consul.NewClient(config)
	if err != nil {
		panic(err)
	}

	s.kv = c.KV()
	p, _, err := s.kv.Get("rpg/latestID", nil)
	if err != nil {
		panic(err)
	}

	s.pIndex, err = strconv.ParseInt(string(p.Value), 10, 64)
	if err != nil {
		panic(err)
	}

	fmt.Printf("[CRM] 最大ID：%d\n", s.pIndex)

	s.mgo, err = mgo.Dial(mgoUrl)
	if err != nil {
		panic(err)
	}

	s.mgo.SetMode(mgo.Monotonic, true)
	fmt.Println("[CRM] 连接 mongo 成功")
}

func (s *CRMService) Signup(c context.Context, in *crm_api.SignupReq, out *crm_api.SignupRsp) error {
	id := bson.NewObjectId()

	out.ID = strconv.FormatInt(s.pIndex, 10)
	out.Token = id.Hex()

	now := time.Now()
	player := &mongo.Player{
		ID:         id,
		DisplayID:  out.ID,
		CreateTime: now,
		UpdateTime: now,
	}

	mc := s.mgo.DB(mongo.DB_GLOBAL).C(mongo.C_PLAYER)
	err := mc.Insert(player)
	if err != nil {
		return err
	}

	s.pIndex += 1
	newUID := strconv.FormatInt(s.pIndex, 10)

	d := &consul.KVPair{Key: "rpg/latestID", Value: []byte(newUID)}
	s.kv.Put(d, nil)

	log.Printf("创建Player: %s - %s\n", id.String(), id.Hex())

	return nil
}

func (s *CRMService) BindPhone(c context.Context, in *crm_api.BindPhoneReq, out *crm_api.BindPhoneRsp) error {
	return nil
}

func (s *CRMService) CRMPing(c context.Context, in *crm_api.CRMPingReq, out *crm_api.CRMPingRsp) error {
	return nil
}