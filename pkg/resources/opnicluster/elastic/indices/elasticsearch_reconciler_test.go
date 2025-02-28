package indices

import (
	"context"
	"log"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	opensearchapiext "github.com/rancher/opni/pkg/resources/opnicluster/elastic/indices/types"
)

var _ = Describe("Indices", Label("unit"), func() {
	var (
		reconciler *elasticsearchReconciler
		transport  *httpmock.MockTransport
	)

	BeforeEach(func() {
		transport = httpmock.NewMockTransport()
		transport.RegisterNoResponder(httpmock.NewNotFoundResponder(nil))
	})

	JustBeforeEach(func() {
		reconciler = newElasticsearchReconcilerWithTransport(context.Background(), "test", transport)
	})
	Context("reconciling ISM policies", func() {
		var (
			policy         opensearchapiext.ISMPolicySpec
			policyResponse *opensearchapiext.ISMGetResponse
		)
		BeforeEach(func() {
			policy = opensearchapiext.ISMPolicySpec{
				ISMPolicyIDSpec: &opensearchapiext.ISMPolicyIDSpec{
					PolicyID:   "testpolicy",
					MarshallID: false,
				},
				Description:  "testing policy",
				DefaultState: "test",
				States: []opensearchapiext.StateSpec{
					{
						Name: "test",
						Actions: []opensearchapiext.ActionSpec{
							{
								ActionOperation: &opensearchapiext.ActionOperation{
									ReadOnly: &opensearchapiext.ReadOnlyOperation{},
								},
							},
						},
						Transitions: make([]opensearchapiext.TransitionSpec, 0),
					},
				},
				ISMTemplate: []opensearchapiext.ISMTemplateSpec{
					{
						IndexPatterns: []string{
							"test*",
						},
						Priority: 100,
					},
				},
			}
			policyResponse = &opensearchapiext.ISMGetResponse{
				ID:          "testid",
				Version:     1,
				PrimaryTerm: 1,
				SeqNo:       1,
				Policy: opensearchapiext.ISMPolicySpec{
					ISMPolicyIDSpec: &opensearchapiext.ISMPolicyIDSpec{
						PolicyID:   "testpolicy",
						MarshallID: true,
					},
					Description:  "testing policy",
					DefaultState: "test",
					States: []opensearchapiext.StateSpec{
						{
							Name: "test",
							Actions: []opensearchapiext.ActionSpec{
								{
									ActionOperation: &opensearchapiext.ActionOperation{
										ReadOnly: &opensearchapiext.ReadOnlyOperation{},
									},
								},
							},
							Transitions: make([]opensearchapiext.TransitionSpec, 0),
						},
					},
					ISMTemplate: []opensearchapiext.ISMTemplateSpec{
						{
							IndexPatterns: []string{
								"test*",
							},
							Priority: 100,
						},
					},
				},
			}
		})
		When("ISM does not exist", func() {
			It("should create a new ISM", func() {
				transport.RegisterResponder(
					"GET",
					"https://opni-es-client.test:9200/_plugins/_ism/policies/testpolicy",
					httpmock.NewStringResponder(404, `{"mesg": "Not found"}`).Once(),
				)
				transport.RegisterResponder(
					"PUT",
					"https://opni-es-client.test:9200/_plugins/_ism/policies/testpolicy",
					httpmock.NewJsonResponderOrPanic(200, policyResponse).Once(),
				)
				Expect(func() error {
					err := reconciler.reconcileISM(policy)
					if err != nil {
						log.Println(err)
					}
					return err
				}()).To(BeNil())
				// Confirm all responders have been called
				Expect(transport.GetTotalCallCount()).To(Equal(transport.NumResponders()))
			})
		})
		When("ISM exists and is the same", func() {
			It("should do nothing", func() {
				transport.RegisterResponder(
					"GET",
					"https://opni-es-client.test:9200/_plugins/_ism/policies/testpolicy",
					httpmock.NewJsonResponderOrPanic(200, policyResponse).Once(),
				)
				Expect(func() error {
					err := reconciler.reconcileISM(policy)
					if err != nil {
						log.Println(err)
					}
					return err
				}()).To(BeNil())
				// Confirm all responders have been called
				Expect(transport.GetTotalCallCount()).To(Equal(transport.NumResponders()))
			})
		})
		When("ISM exists and is different", func() {
			It("should update the policy", func() {
				policyResponseNew := &opensearchapiext.ISMGetResponse{
					ID:          "testid",
					Version:     1,
					PrimaryTerm: 1,
					SeqNo:       2,
					Policy: opensearchapiext.ISMPolicySpec{
						ISMPolicyIDSpec: &opensearchapiext.ISMPolicyIDSpec{
							PolicyID:   "testpolicy",
							MarshallID: true,
						},
						Description:  "testing policy",
						DefaultState: "test",
						States: []opensearchapiext.StateSpec{
							{
								Name: "test",
								Actions: []opensearchapiext.ActionSpec{
									{
										ActionOperation: &opensearchapiext.ActionOperation{
											ReadOnly: &opensearchapiext.ReadOnlyOperation{},
										},
									},
								},
								Transitions: make([]opensearchapiext.TransitionSpec, 0),
							},
						},
						ISMTemplate: []opensearchapiext.ISMTemplateSpec{
							{
								IndexPatterns: []string{
									"test*",
								},
								Priority: 100,
							},
						},
					},
				}
				policy.Description = "this is a different description"
				transport.RegisterResponder(
					"GET",
					"https://opni-es-client.test:9200/_plugins/_ism/policies/testpolicy",
					httpmock.NewJsonResponderOrPanic(200, policyResponse).Once(),
				)
				transport.RegisterResponderWithQuery(
					"PUT",
					"https://opni-es-client.test:9200/_plugins/_ism/policies/testpolicy",
					map[string]string{
						"if_seq_no":       "1",
						"if_primary_term": "1",
					},
					httpmock.NewJsonResponderOrPanic(200, policyResponseNew).Once(),
				)
				Expect(func() error {
					err := reconciler.reconcileISM(policy)
					if err != nil {
						log.Println(err)
					}
					return err
				}()).To(BeNil())
				// Confirm all responders have been called
				Expect(transport.GetTotalCallCount()).To(Equal(transport.NumResponders()))
			})
		})
	})
	Context("reconciling index templates", func() {
		var indexTemplate *opensearchapiext.IndexTemplateSpec
		BeforeEach(func() {
			indexTemplate = &opensearchapiext.IndexTemplateSpec{
				TemplateName: "testtemplate",
				IndexPatterns: []string{
					"test*",
				},
				Template: opensearchapiext.TemplateSpec{
					Settings: opensearchapiext.TemplateSettingsSpec{
						NumberOfShards:   1,
						NumberOfReplicas: 1,
						ISMPolicyID:      logPolicyName,
						RolloverAlias:    logIndexAlias,
					},
					Mappings: opensearchapiext.TemplateMappingsSpec{
						Properties: map[string]opensearchapiext.PropertySettings{
							"timestamp": {
								Type: "date",
							},
						},
					},
				},
			}
		})
		When("the template does not exist", func() {
			It("should create the index template", func() {
				transport.RegisterResponder(
					"GET",
					"https://opni-es-client.test:9200/_index_template/testtemplate",
					httpmock.NewStringResponder(404, `{"mesg": "Not found"}`).Once(),
				)
				transport.RegisterResponderWithQuery(
					"PUT",
					"https://opni-es-client.test:9200/_index_template/testtemplate",
					map[string]string{
						"create": "true",
					},
					httpmock.NewStringResponder(200, `{"status": "complete"}`).Once(),
				)
				Expect(func() error {
					err := reconciler.maybeCreateIndexTemplate(indexTemplate)
					if err != nil {
						log.Println(err)
					}
					return err
				}()).To(BeNil())
				// Confirm all responders have been called
				Expect(transport.GetTotalCallCount()).To(Equal(transport.NumResponders()))
			})
		})
		When("the template does exist", func() {
			It("should do nothing", func() {
				transport.RegisterResponder(
					"GET",
					"https://opni-es-client.test:9200/_index_template/testtemplate",
					httpmock.NewStringResponder(200, `{"mesg": "found it"}`).Once(),
				)
				Expect(func() error {
					err := reconciler.maybeCreateIndexTemplate(indexTemplate)
					if err != nil {
						log.Println(err)
					}
					return err
				}()).To(BeNil())
				// Confirm all responders have been called
				Expect(transport.GetTotalCallCount()).To(Equal(transport.NumResponders()))
			})
		})
	})
	Context("reconciling rollover indices", func() {
		var (
			prefix = "test"
			alias  = "test"
		)
		When("alias index, and rollover indices don't exist", func() {
			It("should bootstrap the index", func() {
				transport.RegisterResponderWithQuery(
					"GET",
					"https://opni-es-client.test:9200/_cat/indices/test*",
					map[string]string{
						"format": "json",
					},
					httpmock.NewStringResponder(200, "[]").Once(),
				)
				transport.RegisterResponderWithQuery(
					"GET",
					"https://opni-es-client.test:9200/_cat/indices/test",
					map[string]string{
						"format": "json",
					},
					httpmock.NewStringResponder(404, `{"mesg": "Not found"}`).Once(),
				)
				transport.RegisterResponder(
					"POST",
					"https://opni-es-client.test:9200/_aliases",
					httpmock.NewStringResponder(200, "OK").Once(),
				)
				transport.RegisterResponder(
					"PUT",
					"https://opni-es-client.test:9200/test-000001",
					httpmock.NewStringResponder(200, "OK").Once(),
				)
				Expect(func() error {
					err := reconciler.maybeBootstrapIndex(prefix, alias)
					if err != nil {
						log.Println(err)
					}
					return err
				}()).To(BeNil())
				// Confirm all responders have been called
				Expect(transport.GetTotalCallCount()).To(Equal(transport.NumResponders()))
			})
		})
		When("alias index exists, and rollover indices don't", func() {
			It("should reindex into the bootstrap index", func() {
				transport.RegisterResponderWithQuery(
					"GET",
					"https://opni-es-client.test:9200/_cat/indices/test*",
					map[string]string{
						"format": "json",
					},
					httpmock.NewStringResponder(200, "[]").Once(),
				)
				transport.RegisterResponderWithQuery(
					"GET",
					"https://opni-es-client.test:9200/_cat/indices/test",
					map[string]string{
						"format": "json",
					},
					httpmock.NewStringResponder(200, `{"status": "exists"}`).Once(),
				)
				transport.RegisterResponder(
					"PUT",
					"https://opni-es-client.test:9200/test-000001",
					httpmock.NewStringResponder(200, "OK").Once(),
				)
				transport.RegisterResponderWithQuery(
					"POST",
					"https://opni-es-client.test:9200/_reindex",
					map[string]string{
						"wait_for_completion": "true",
					},
					httpmock.NewStringResponder(200, `{"status": "OK"}`).Once(),
				)
				transport.RegisterResponder(
					"POST",
					"https://opni-es-client.test:9200/_aliases",
					httpmock.NewStringResponder(200, "OK").Once(),
				)
				Expect(func() error {
					err := reconciler.maybeBootstrapIndex(prefix, alias)
					if err != nil {
						log.Println(err)
					}
					return err
				}()).To(BeNil())
				// Confirm all responders have been called
				Expect(transport.GetTotalCallCount()).To(Equal(transport.NumResponders()))
			})
		})
		When("rollover indices exist", func() {
			It("should do nothing", func() {
				transport.RegisterResponderWithQuery(
					"GET",
					"https://opni-es-client.test:9200/_cat/indices/test*",
					map[string]string{
						"format": "json",
					},
					httpmock.NewStringResponder(200, `[{"test-000002": "thisexists"}, {"test-000003": "this also exists"}]`).Once(),
				)
				Expect(func() error {
					err := reconciler.maybeBootstrapIndex(prefix, alias)
					if err != nil {
						log.Println(err)
					}
					return err
				}()).To(BeNil())
				// Confirm all responders have been called
				Expect(transport.GetTotalCallCount()).To(Equal(transport.NumResponders()))
			})
		})
	})
	Context("reconciling indices", func() {
		var (
			indexName     string
			indexSettings = map[string]opensearchapiext.TemplateMappingsSpec{}
		)
		BeforeEach(func() {
			indexName = "test"
			indexSettings = map[string]opensearchapiext.TemplateMappingsSpec{
				"mappings": {
					Properties: map[string]opensearchapiext.PropertySettings{
						"start_ts": {
							Type:   "date",
							Format: "epoch_millis",
						},
						"end_ts": {
							Type:   "date",
							Format: "epoch_millis",
						},
					},
				},
			}
		})
		When("index does not exist", func() {
			It("should create the index settings", func() {
				transport.RegisterResponderWithQuery(
					"GET",
					"https://opni-es-client.test:9200/_cat/indices/test",
					map[string]string{
						"format": "json",
					},
					httpmock.NewStringResponder(404, `{"mesg": "Not found"}`).Once(),
				)
				transport.RegisterResponder(
					"PUT",
					"https://opni-es-client.test:9200/test",
					httpmock.NewStringResponder(200, "OK").Once(),
				)
				Expect(func() error {
					err := reconciler.maybeCreateIndex(indexName, indexSettings)
					if err != nil {
						log.Println(err)
					}
					return err
				}()).To(BeNil())
				// Confirm all responders have been called
				Expect(transport.GetTotalCallCount()).To(Equal(transport.NumResponders()))
			})
		})
		When("index does exist", func() {
			It("should do nothing", func() {
				transport.RegisterResponderWithQuery(
					"GET",
					"https://opni-es-client.test:9200/_cat/indices/test",
					map[string]string{
						"format": "json",
					},
					httpmock.NewStringResponder(200, "OK").Once(),
				)
				Expect(func() error {
					err := reconciler.maybeCreateIndex(indexName, indexSettings)
					if err != nil {
						log.Println(err)
					}
					return err
				}()).To(BeNil())
				// Confirm all responders have been called
				Expect(transport.GetTotalCallCount()).To(Equal(transport.NumResponders()))
			})
		})
	})
	Context("reconciling kibana objects", func() {
		When("kibana tracking index doesn't exist", func() {
			It("should create the tracking index and the kibana objects", func() {
				transport.RegisterResponderWithQuery(
					"GET",
					"https://opni-es-client.test:9200/_cat/indices/opni-dashboard-version",
					map[string]string{
						"format": "json",
					},
					httpmock.NewStringResponder(404, `{"mesg": "Not found"}`).Once(),
				)
				transport.RegisterResponderWithQuery(
					"POST",
					"http://admin:admin@opni-es-kibana.test:5601/api/saved_objects/_import",
					map[string]string{
						"overwrite": "true",
					},
					httpmock.NewStringResponder(200, "OK").Once(),
				)
				transport.RegisterResponder(
					"POST",
					"https://opni-es-client.test:9200/opni-dashboard-version/_doc/latest/_update",
					httpmock.NewStringResponder(200, "OK").Once(),
				)
				Expect(func() error {
					err := reconciler.importKibanaObjects()
					if err != nil {
						log.Println(err)
					}
					return err
				}()).To(BeNil())
				// Confirm all responders have been called
				Expect(transport.GetTotalCallCount()).To(Equal(transport.NumResponders()))
			})
		})
		When("kibana tracking index has version that is old", func() {
			It("should update the tracking index and the kibana objects", func() {
				kibanaResponse := opensearchapiext.KibanaDocResponse{
					Index:       "opni-dashboard-version",
					ID:          "latest",
					SeqNo:       1,
					PrimaryTerm: 1,
					Found:       opensearchapi.BoolPtr(true),
					Source: opensearchapiext.KibanaVersionDoc{
						DashboardVersion: "0.0.1",
					},
				}
				transport.RegisterResponderWithQuery(
					"GET",
					"https://opni-es-client.test:9200/_cat/indices/opni-dashboard-version",
					map[string]string{
						"format": "json",
					},
					httpmock.NewStringResponder(200, "OK").Once(),
				)
				transport.RegisterResponder(
					"GET",
					"https://opni-es-client.test:9200/opni-dashboard-version/_doc/latest",
					httpmock.NewJsonResponderOrPanic(200, kibanaResponse).Once(),
				)
				transport.RegisterResponderWithQuery(
					"POST",
					"http://admin:admin@opni-es-kibana.test:5601/api/saved_objects/_import",
					map[string]string{
						"overwrite": "true",
					},
					httpmock.NewStringResponder(200, "OK").Once(),
				)
				transport.RegisterResponder(
					"POST",
					"https://opni-es-client.test:9200/opni-dashboard-version/_doc/latest/_update",
					httpmock.NewStringResponder(200, "OK").Once(),
				)
				Expect(func() error {
					err := reconciler.importKibanaObjects()
					if err != nil {
						log.Println(err)
					}
					return err
				}()).To(BeNil())
				// Confirm all responders have been called
				Expect(transport.GetTotalCallCount()).To(Equal(transport.NumResponders()))
			})
		})
		When("kibana tracking index exists and is up to date", func() {
			It("should do nothing", func() {
				kibanaResponse := opensearchapiext.KibanaDocResponse{
					Index:       "opni-dashboard-version",
					ID:          "latest",
					SeqNo:       1,
					PrimaryTerm: 1,
					Found:       opensearchapi.BoolPtr(true),
					Source: opensearchapiext.KibanaVersionDoc{
						DashboardVersion: kibanaDashboardVersion,
					},
				}
				transport.RegisterResponderWithQuery(
					"GET",
					"https://opni-es-client.test:9200/_cat/indices/opni-dashboard-version",
					map[string]string{
						"format": "json",
					},
					httpmock.NewStringResponder(200, "OK").Once(),
				)
				transport.RegisterResponder(
					"GET",
					"https://opni-es-client.test:9200/opni-dashboard-version/_doc/latest",
					httpmock.NewJsonResponderOrPanic(200, kibanaResponse).Once(),
				)
				Expect(func() error {
					err := reconciler.importKibanaObjects()
					if err != nil {
						log.Println(err)
					}
					return err
				}()).To(BeNil())
				// Confirm all responders have been called
				Expect(transport.GetTotalCallCount()).To(Equal(transport.NumResponders()))
			})
		})
	})
})
