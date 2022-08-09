package util

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/proto"
)

func ESCheckAndRepair(cfg *proto.ConfigNew, EsWrite escli.ESClient) error {
	if cfg.EsVersion == 6 {
		return es6CheckAndRepair(cfg, EsWrite)
	}
	if cfg.EsVersion == 7 {
		return es7CheckAndRepair(cfg, EsWrite)
	}
	return nil
}

func es6CheckAndRepair(cfg *proto.ConfigNew, EsWrite escli.ESClient) error {
	// curl -X PUT http://localhost:9200/${prefix}proof_config/proof_config/member-${namager1}/_create -d'{"address":"'${namager1}'","role": "manager","organization": "system", "note" : "system-init-manager"}' -H "Content-Type: application/json"
	for _, v := range cfg.ManagerAddress {
		value := fmt.Sprintf(`{"address":"%v","role":"manager","organization":"system","note":"system-init-manager"}`, v)
		err := EsWrite.Update("proof_config", "proof_config", fmt.Sprintf("member-%v", v), value)
		if err != nil {
			return err
		}
	}

	// curl -X PUT http://localhost:9200/${prefix}proof_config_org/proof_config_org/org-system/_create -d'{"count": '${namagerCount}',"organization": "system", "note" : "system-init"}' -H "Content-Type: application/json"
	value := fmt.Sprintf(`{"count": %v,"organization": "system", "note" : "system-init"}`, len(cfg.ManagerAddress))
	err := EsWrite.Update("proof_config_org", "proof_config_org", "org-system", value)
	if err != nil {
		return err
	}

	// curl -X PUT http://localhost:9200/${prefix}proof/proof/proof-create-proof-index/_create -d '{"init": 1}' -H "Content-Type: application/json"
	err = EsWrite.Update("proof", "proof", "proof-create-proof-index", `{"init": 1}`)
	if err != nil {
		return err
	}

	// curl -X PUT http://localhost:9200/${prefix}proof/proof/_mapping?pretty -d '{"proof": {"properties": {"捐赠数量": {"type": "double"}}}}' -H "Content-Type: application/json"
	err = EsWrite.Update("proof", "proof", "_mapping", `{"proof": {"properties": {"捐赠数量": {"type": "double"}}}}`)
	if err != nil {
		return err
	}

	// curl -X PUT http://localhost:9200/${prefix}proof/proof/_mapping?pretty -d '{"proof": {"properties": {"志愿者数量": {"type": "long"}}}}' -H "Content-Type: application/json"
	err = EsWrite.Update("proof", "proof", "_mapping", `{"proof": {"properties": {"志愿者数量": {"type": "long"}}}}`)
	if err != nil {
		return err
	}

	// curl -X PUT http://localhost:9200/${prefix}proof/_settings -d '{ "index" : { "max_result_window" : 20000}}' -H "Content-Type: application/json"
	res, err := EsWrite.PerformRequest("PUT", fmt.Sprintf("/%vproof/_settings", cfg.ConvertEs.Prefix), url.Values{},
		`{ "index" : { "max_result_window" : 20000}}`,
		http.Header{})
	if err != nil {
		return err
	}

	// curl -X PUT  http://localhost:9200/${prefix}proof/_settings -d '{"index.mapping.total_fields.limit":1000000}' -H "Content-Type: application/json"
	res, err = EsWrite.PerformRequest("PUT", fmt.Sprintf("/%vproof/_settings", cfg.ConvertEs.Prefix), url.Values{},
		`{"index.mapping.total_fields.limit":1000000}`,
		http.Header{})
	if err != nil {
		return err
	}

	// curl -X DELETE http://localhost:9200/${prefix}proof/proof/proof-create-proof-index -H "Content-Type: application/json"
	res, err = EsWrite.PerformRequest("DELETE", fmt.Sprintf("/%vproof/proof/proof-create-proof-index", cfg.ConvertEs.Prefix), url.Values{},
		nil,
		http.Header{})
	if err != nil {
		return err
	}
	log.Info("es 自动处理数据成功", "res", res)
	return nil
}

func es7CheckAndRepair(cfg *proto.ConfigNew, EsWrite escli.ESClient) error {
	// curl -X PUT http://localhost:9200/${prefix}proof_config -d '{"settings":{"number_of_shards":'${numberOfShards}',"number_of_replicas":'${numberOfReplicas}'}}' -H "Content-Type: application/json"
	exists, err := EsWrite.Exists("proof_config")
	if err != nil {
		return err
	}
	if !exists {
		_, err := EsWrite.PerformRequest("PUT", fmt.Sprintf("/%vproof_config", cfg.ConvertEs.Prefix), url.Values{},
			fmt.Sprintf(`{"settings":{"number_of_shards": %v,"number_of_replicas":%v}}`, cfg.EsIndex.NumberOfShards, cfg.EsIndex.NumberOfReplicas),
			http.Header{})
		if err != nil {
			return err
		}

	}

	// curl -X PUT http://localhost:9200/${prefix}proof_config/_doc/member-${manager1}/_create  -d'{"address":"'${manager1}'","role": "manager","organization": "system", "note" : "system-init-manager"}' -H "Content-Type: application/json"
	for _, v := range cfg.ManagerAddress {
		value := fmt.Sprintf(`{"address":"%v","role":"manager","organization":"system","note":"system-init-manager"}`, v)
		err := EsWrite.Update("proof_config", "_doc", fmt.Sprintf("member-%v", v), value)
		if err != nil {
			return err
		}
	}

	// curl -X PUT http://localhost:9200/${prefix}proof_config_org -d'{"settings":{"number_of_shards":'${numberOfShards}',"number_of_replicas":'${numberOfReplicas}'}}' -H "Content-Type: application/json"
	exists, err = EsWrite.Exists("proof_config_org")
	if err != nil {
		return err
	}
	if !exists {
		_, err := EsWrite.PerformRequest("PUT", fmt.Sprintf("/%vproof_config_org", cfg.ConvertEs.Prefix), url.Values{},
			fmt.Sprintf(`{"settings":{"number_of_shards": %v,"number_of_replicas":%v}}`, cfg.EsIndex.NumberOfShards, cfg.EsIndex.NumberOfReplicas),
			http.Header{})
		if err != nil {
			return err
		}
	}

	// curl -X PUT http://localhost:9200/${prefix}proof_config_org/_doc/org-system/_create  -d'{"count": '${managerCount}',"organization": "system", "note" : "system-init"}' -H "Content-Type: application/json"
	err = EsWrite.Update("proof_config_org", "_doc", "org-system",
		fmt.Sprintf(`{"count": %v,"organization": "system", "note" : "system-init"}`, len(cfg.ManagerAddress)))
	if err != nil {
		return err
	}

	// curl -X PUT  http://localhost:9200/${prefix}proof -d '{"settings":{"number_of_shards":'${numberOfShards}',"number_of_replicas":'${numberOfReplicas}'}}' -H "Content-Type: application/json"
	exists, err = EsWrite.Exists("proof")
	if err != nil {
		return err
	}
	if !exists {
		_, err := EsWrite.PerformRequest("PUT", fmt.Sprintf("/%vproof", cfg.ConvertEs.Prefix), url.Values{},
			fmt.Sprintf(`{"settings":{"number_of_shards": %v,"number_of_replicas":%v}}`, cfg.EsIndex.NumberOfShards, cfg.EsIndex.NumberOfReplicas),
			http.Header{})
		if err != nil {
			return err
		}
	}

	// curl -X PUT http://localhost:9200/${prefix}proof/_doc/proof-create-proof-index/_create -d '{"init": 1}' -H "Content-Type: application/json"
	err = EsWrite.Update("proof", "_doc", "proof-create-proof-index", `{"init": 1}`)
	if err != nil {
		return err
	}

	// curl -X PUT http://localhost:9200/${prefix}proof/_mapping?pretty -d '{"properties": {"捐赠数量": {"type": "double"}}}' -H "Content-Type: application/json"
	// curl -X PUT http://localhost:9200/${prefix}proof/_mapping?pretty -d '{"properties": {"志愿者数量": {"type": "long"}}}' -H "Content-Type: application/json"
	_, err = EsWrite.PerformRequest("PUT", fmt.Sprintf("/%vproof/_mapping", cfg.ConvertEs.Prefix), url.Values{},
		`{"properties": {"捐赠数量": {"type": "double"}, "志愿者数量": {"type": "long"}}}`,
		http.Header{})
	if err != nil {
		return err
	}

	// curl -X PUT http://localhost:9200/${prefix}proof/_settings -d '{"index":{"max_result_window":20000}}' -H "Content-Type: application/json"
	// curl -X PUT  http://localhost:9200/${prefix}proof/_settings -d '{"index.mapping.total_fields.limit":1000000}' -H "Content-Type: application/json"
	_, err = EsWrite.PerformRequest("PUT", fmt.Sprintf("/%vproof/_settings", cfg.ConvertEs.Prefix), url.Values{},
		`{"index.max_result_window":20000, "index.mapping.total_fields.limit":1000000}`,
		http.Header{})
	if err != nil {
		return err
	}

	// curl -X DELETE http://localhost:9200/${prefix}proof/_doc/proof-create-proof-index -H "Content-Type: application/json"
	_, err = EsWrite.PerformRequest("DELETE", fmt.Sprintf("/%vproof/_doc/proof-create-proof-index", cfg.ConvertEs.Prefix), url.Values{},
		nil,
		http.Header{})
	if err != nil {
		return err
	}
	return nil
}
