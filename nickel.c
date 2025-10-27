#include <stdint.h>

extern uintptr_t traceCallback(void*, uint8_t*, uintptr_t);

uintptr_t traceCallbackTrampoline(void* context, const uint8_t* buf, uintptr_t len) {
	// Yeah, we're casting away the const. Go doesn't know about const pointers,
	// so traceCallback can't accept one. But we promise not to write to it.
	return traceCallback(context, (uint8_t*)buf, len);
}
