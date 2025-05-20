package grpc_metadata

import "google.golang.org/grpc/metadata"

func MDToMapString(md metadata.MD) map[string]string {
	var mdMapString map[string]string = map[string]string{}

	for k, v := range md {
		if len(v) <= 0 {
			continue
		}
		mdMapString[k] = v[0]
	}

	return mdMapString
}

func MDGetString(md metadata.MD, key string) string {
	var values []string = md.Get(key)

	if len(values) <= 0 {
		return ""
	}

	return values[0]
}
