package main

import (
	"flag"
	"time"

	"github.com/antoninbas/go-ipfix/pkg/ipfix"
	"k8s.io/klog"
)

func main() {
	klog.InitFlags(nil)

	host := flag.String("host", "127.0.0.1", "Address of collector")
	port := flag.Int("port", 4739, "Port of collector")
	count := flag.Int("count", 0, "Number of messages to send, 0 means forever (1s interval)")
	flag.Parse()

	if err := ipfix.Init(); err != nil {
		klog.Fatalf("Initialization error: %v", err)
	}
	defer ipfix.Cleanup()

	ipf, err := ipfix.NewIPFix(12345)
	if err != nil {
		klog.Errorf("Error when opening exporter: %v", err)
		return
	}
	defer ipf.Close()

	klog.Infof("Exporter opened successfully")

	if err := ipf.AddCollector(*host, *port, ipfix.IPFixProtoTCP); err != nil {
		klog.Errorf("Error when adding collector: %v", err)
		return
	}

	klog.Infof("Collector added successfully")

	tpl, err := ipf.NewDataTemplate(2)
	if err != nil {
		klog.Errorf("Error when creating template: %v", err)
	}

	if err := ipf.AddField(tpl, ipfix.IPFIX_ENO_Default, ipfix.IPFIX_FIELD_sourceIPv4Address, 4); err != nil {
		klog.Errorf("Error when adding sourceIPv4Address: %v", err)
		return
	}
	if err := ipf.AddField(tpl, ipfix.IPFIX_ENO_Default, ipfix.IPFIX_FIELD_packetDeltaCount, 8); err != nil {
		klog.Errorf("Error when adding packetDeltaCount: %v", err)
		return
	}

	i := 0
	for {
		if *count > 0 && i >= *count {
			break
		}
		i++
		fArray := ipfix.NewFieldArray(tpl, 2)
		fArray.AddU32(0x11223344)
		fArray.AddU64(10000)
		klog.Infof("Sending message %d", i)
		if err := ipf.Export(fArray); err != nil {
			klog.Errorf("Export failed: %v", err)
		}
		fArray.Delete()
		time.Sleep(time.Second)
	}
	klog.Infof("DONE!")
}
