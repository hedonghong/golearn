package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/olivere/elastic/v7"
	"log"
	"os"
	"time"
)

var (
	host = "http://127.0.0.1:9200"
	ElasticClient *elastic.Client
)

//1、SetHttpClient(*http.Client)允许您配置自己的http.Client和/或http.Transport（默认为http.DefaultClient）；在许多弹性实例中使用相同的http.Client（即使使用http.DefaultClient）是一个好主意，以便有效地使用打开的TCP连接。
//
//2、StURURL（…字符串）允许您指定要连接的URL（默认值是http://127.0.0.1:9200）。
//
//3、StasBaseCuthe（用户名，密码字符串）允许您指定HTTP基本身份验证详细信息。使用这个，例如用盾牌。
//
//4、SETSNIFF（BOOL）允许您指定弹性是否应该定期检查集群（默认为真）。
//
//5、StSnIFFEffTimeOUT（时间。持续时间）是嗅探节点弹出时间之前的时间（默认为2秒）。
//
//6、StnSnFiffer-TimeOutExpLoT（时间。持续时间）是创建新客户端时使用的嗅探超时。它通常比嗅探器超时大，并且证明对慢启动有帮助（默认为5秒）。
//
//7、StnSnIFFER间隔（时间。持续时间）允许您指定两个嗅探器进程之间的间隔（默认为15分钟）。
//
//8、SetHealthcheck（bool）允许您通过尝试定期连接到它的节点（默认为true）来指定Elastic是否将执行健康检查。
//
//9、SethalthCuffTimeExt（时间。持续时间）是健康检查的超时时间（默认值为1秒）。
//
//10、SethalthCuffTimeOutExtudio（时间。持续时间）是创建新客户端时使用的健康检查超时。它通常大于健康检查超时，并可能有助于慢启动（默认为5秒）。
//
//11、sethealthcheckinterval（time.duration）指定间隔之间的两个健康检查（默认是60秒）。
//
//12、
//SetDecoder（.ic.Decoder）允许您为来自Elasticsearch的JSON消息设置自己的解码器（默认为&.ic.DefaultDecoder{}）。
//
//13、StError日志（*Log.LoggER）将日志记录器设置为用于错误消息（默认为NIL）。错误日志将包含例如关于加入群集的节点或标记为“死亡”的消息。
//
//14、
//SETIN FLUOG（*Log.LoggER）将记录器设置为用于信息性消息（默认为NIL）。信息日志将包含例如请求和它们的响应时间。
//15、
//StReTraceLoG（*Log.LoggER）设置用于打印HTTP请求和响应（默认为NIL）的记录器。这有助于调试有线上正在发生的事情
//
//16、
//StestRealdPuelin（插件…字符串）设置需要注册的插件列表。弹性将设法在启动时找到它们。如果没有找到其中一个，则在启动时会发现一个类型的弹性错误。
//
//17、
//StReReTrice（…）设置用于处理失败请求的重试策略。详情请参阅重试和退避
//
//18、
//SETGZIP（BOOL）启用或禁用请求端的压缩。默认情况下禁用。

func init()  {
	client, err := elastic.NewClient(
		elastic.SetURL(host),
		elastic.SetSniff(false),
		elastic.SetHealthcheckInterval(10 * time.Second),
		elastic.SetGzip(true),
		elastic.SetErrorLog(log.New(os.Stderr, "ELASTIC", log.LstdFlags)),
		elastic.SetInfoLog(log.New(os.Stdout, "", log.LstdFlags)),
		)
	if err != nil {
		panic(err)
	}
	ElasticClient = client
}

func main()  {
	fmt.Println(ElasticClient)
	mapping := `
	{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0
		},
		"mappings":{
			"properties": {
				"id" : {"type":"long"},
				"name": {"type":"keyword"},
				"first_letter": {"type":"keyword"},
				"sort": {"type":"integer"},
				"factory_status": {"type":"integer"},
				"show_status": {"type":"integer"},
				"product_comme": {"type":"integer"},
				"logo": {"type":"text"},
				"big_pic": {"type":"text"},
				"brand_story": {"type":"text"}
			}
		}
	}
`
	indexName := "pms_brand"
	ctx := context.Background()

	pingResult, code, err := ElasticClient.Ping(host).Do(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println("elasticsearch ", code, pingResult.Version.Number)

	exists, err := ElasticClient.IndexExists(indexName).Do(ctx)
	if err != nil {
		panic(err)
	}
	if !exists {
		createIndex, err := ElasticClient.CreateIndex(indexName).BodyString(mapping).Do(ctx)
		if err != nil {
			panic(err)
		}
		if !createIndex.Acknowledged {
			errors.New("CreateIndex:"+indexName+", no Acknowledged")
		}
	}

//	brand := `{"id": 1, "name":"", "first_letter":"",
//"sort": 1, "factory_status":1, "show_status":1, "product_comme": 1,
//"log":"http://macro-oss.oss-cn-shenzhen.aliyuncs.com/mall/images/20180607/timg(5).jpg",
//"big_pic":"", "brand_story":"Victoria's Secret的故事"
//}`
//	put, err := ElasticClient.Index().Index(indexName).BodyString(brand).Do(ctx)
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println("put:", put.Id, put.Index, put.Type)

	//ElasticClient.Delete().Index(indexName).Id("WIfrY3sBU2YibaLuG7z2").Do(ctx)

	//termQuery := elastic.NewTermQuery("name", "mkie")
	//ElasticClient.DeleteByQuery().Index(indexName).Query(termQuery).Do(ctx)

	//ElasticClient.Flush(indexName).Do(ctx)

	get, err := ElasticClient.Get().Index(indexName).Id("WIfrY3sBU2YibaLuG7z2").Do(ctx)
	if err != nil {
		panic(err)
	}
	if get.Found {
		fmt.Println("document:", get)
	}

	//ElasticClient.DeleteIndex(indexName).Do(ctx)

	//bulkRequest := ElasticClient.Bulk()
	//req1 := elastic.NewBulkIndexRequest().Index(indexName).Id("xx").Doc("")
	//bulkRequest = bulkRequest.Add(req1)
	//bulkRequest.Do(ctx)
}
