//go:build darwin

package main

/*
#cgo pkg-config: libusb-1.0
#include <libusb.h>
#include <stdlib.h>

#define POS_ERR_NODEV -1001
#define POS_ERR_NOEP  -1002

typedef struct {
	libusb_context *ctx;
	libusb_device_handle *handle;
	unsigned char ep;
	int iface;
	int err;
} pos_printer;

static int find_out_endpoint(libusb_device *dev, unsigned char *ep, int *iface) {
	struct libusb_config_descriptor *cfg = NULL;
	int rc = libusb_get_active_config_descriptor(dev, &cfg);
	if (rc != 0 || cfg == NULL) {
		return rc == 0 ? POS_ERR_NOEP : rc;
	}
	for (int i = 0; i < cfg->bNumInterfaces; i++) {
		const struct libusb_interface *intf = &cfg->interface[i];
		for (int a = 0; a < intf->num_altsetting; a++) {
			const struct libusb_interface_descriptor *alt = &intf->altsetting[a];
			if (alt->bInterfaceClass != LIBUSB_CLASS_PRINTER) {
				continue;
			}
			for (int e = 0; e < alt->bNumEndpoints; e++) {
				const struct libusb_endpoint_descriptor *ed = &alt->endpoint[e];
				if ((ed->bmAttributes & LIBUSB_TRANSFER_TYPE_MASK) == LIBUSB_TRANSFER_TYPE_BULK &&
					(ed->bEndpointAddress & LIBUSB_ENDPOINT_DIR_MASK) == LIBUSB_ENDPOINT_OUT) {
					*ep = ed->bEndpointAddress;
					*iface = alt->bInterfaceNumber;
					libusb_free_config_descriptor(cfg);
					return 0;
				}
			}
		}
	}
	libusb_free_config_descriptor(cfg);
	return POS_ERR_NOEP;
}

pos_printer pos_open(void) {
	pos_printer out;
	out.ctx = NULL;
	out.handle = NULL;
	out.ep = 0;
	out.iface = 0;
	out.err = 0;

	int rc = libusb_init(&out.ctx);
	if (rc != 0) {
		out.err = rc;
		return out;
	}

	libusb_device **list = NULL;
	ssize_t n = libusb_get_device_list(out.ctx, &list);
	if (n < 0) {
		out.err = (int)n;
		libusb_exit(out.ctx);
		out.ctx = NULL;
		return out;
	}

	libusb_device *chosen = NULL;
	unsigned char chosen_ep = 0;
	int chosen_iface = 0;
	libusb_device *fallback = NULL;
	unsigned char fallback_ep = 0;
	int fallback_iface = 0;

	for (ssize_t i = 0; i < n; i++) {
		struct libusb_device_descriptor desc;
		if (libusb_get_device_descriptor(list[i], &desc) != 0) {
			continue;
		}
		unsigned char ep = 0;
		int iface = 0;
		if (find_out_endpoint(list[i], &ep, &iface) != 0) {
			continue;
		}
		if (desc.idVendor == 0x0fe6 && desc.idProduct == 0x811e) {
			chosen = list[i];
			chosen_ep = ep;
			chosen_iface = iface;
			break;
		}
		if (fallback == NULL) {
			fallback = list[i];
			fallback_ep = ep;
			fallback_iface = iface;
		}
	}
	if (chosen == NULL) {
		chosen = fallback;
		chosen_ep = fallback_ep;
		chosen_iface = fallback_iface;
	}
	if (chosen == NULL) {
		libusb_free_device_list(list, 1);
		libusb_exit(out.ctx);
		out.ctx = NULL;
		out.err = POS_ERR_NODEV;
		return out;
	}

	rc = libusb_open(chosen, &out.handle);
	libusb_free_device_list(list, 1);
	if (rc != 0) {
		libusb_exit(out.ctx);
		out.ctx = NULL;
		out.handle = NULL;
		out.err = rc;
		return out;
	}

	libusb_set_auto_detach_kernel_driver(out.handle, 1);
	rc = libusb_claim_interface(out.handle, chosen_iface);
	if (rc != 0) {
		libusb_close(out.handle);
		libusb_exit(out.ctx);
		out.ctx = NULL;
		out.handle = NULL;
		out.err = rc;
		return out;
	}
	out.ep = chosen_ep;
	out.iface = chosen_iface;
	return out;
}

int pos_write(libusb_device_handle *handle, unsigned char ep, unsigned char *data, int len) {
	int off = 0;
	while (off < len) {
		int xfer = 0;
		int chunk = len - off;
		if (chunk > 4096) {
			chunk = 4096;
		}
		int rc = libusb_bulk_transfer(handle, ep, data + off, chunk, &xfer, 8000);
		if (rc != 0) {
			return rc;
		}
		if (xfer <= 0) {
			return LIBUSB_ERROR_IO;
		}
		off += xfer;
	}
	return 0;
}

void pos_close(pos_printer p) {
	if (p.handle != NULL) {
		libusb_release_interface(p.handle, p.iface);
		libusb_close(p.handle);
	}
	if (p.ctx != NULL) {
		libusb_exit(p.ctx);
	}
}
*/
import "C"
import (
	"fmt"
	"io"
	"sync"
	"unsafe"
)

var printMu sync.Mutex

type usbPrinter struct {
	dev     C.pos_printer
	closed  bool
	release func()
}

func openPrinter(printerName string) (io.WriteCloser, error) {
	printMu.Lock()
	dev := C.pos_open()
	if dev.err != 0 {
		printMu.Unlock()
		return nil, fmt.Errorf("%s (printer_name %s)", usbErr(dev.err), printerName)
	}
	return &usbPrinter{dev: dev, release: printMu.Unlock}, nil
}

func (p *usbPrinter) Write(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}
	rc := C.pos_write(p.dev.handle, p.dev.ep, (*C.uchar)(unsafe.Pointer(&buf[0])), C.int(len(buf)))
	if rc != 0 {
		return 0, fmt.Errorf("%s", usbErr(rc))
	}
	return len(buf), nil
}

func (p *usbPrinter) Close() error {
	if p.closed {
		return nil
	}
	p.closed = true
	C.pos_close(p.dev)
	if p.release != nil {
		p.release()
	}
	return nil
}

func usbErr(code C.int) string {
	switch code {
	case -1001:
		return "printer USB tidak ditemukan"
	case -1002:
		return "endpoint cetak USB tidak ditemukan"
	default:
		return C.GoString(C.libusb_error_name(code))
	}
}
