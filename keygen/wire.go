package keygen

// -------- license create

type licenseCreateRequest struct {
	Data licenseCreateData `json:"data"`
}

type licenseCreateData struct {
	Type          string                     `json:"type"`
	Attributes    licenseCreateAttributes    `json:"attributes"`
	Relationships licenseCreateRelationships `json:"relationships"`
}

type licenseCreateAttributes struct {
	Expiry   *string         `json:"expiry,omitempty"`
	Metadata LicenseMetadata `json:"metadata,omitempty"`
}

type licenseCreateRelationships struct {
	Policy licenseRelationship `json:"policy"`
}

type licenseRelationship struct {
	Data relationshipData `json:"data"`
}

type relationshipData struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type licenseCreateResponse struct {
	Data struct {
		Attributes struct {
			Key string `json:"key"`
		} `json:"attributes"`
	} `json:"data"`
}

// -------- get license by subscription

type getLicenseBySubscriptionResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// -------- list by policy (rich)

type listLicensesByPolicyResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes struct {
			Key      string         `json:"key,omitempty"`
			Status   string         `json:"status,omitempty"`
			Metadata map[string]any `json:"metadata,omitempty"`
		} `json:"attributes"`
	} `json:"data"`
	Meta struct {
		Page struct {
			CurrentPage int `json:"current"`
			TotalPages  int `json:"total"`
		} `json:"page"`
	} `json:"meta"`
}

// -------- validate

type validateLicenseRequest struct {
	Meta validateMeta `json:"meta"`
}

type validateMeta struct {
	Key   string           `json:"key"`
	Scope fingerprintScope `json:"scope"`
}

type fingerprintScope struct {
	Fingerprint string `json:"fingerprint"`
}

type resolveLicenseIDRequest struct {
	Meta struct {
		Key string `json:"key"`
	} `json:"meta"`
}

type licenseValidationResponse struct {
	Meta struct {
		Valid     bool   `json:"valid"`
		Code      string `json:"code"`
		Detail    string `json:"detail"`
		Timestamp string `json:"ts"`
		Scope     struct {
			Fingerprint string `json:"fingerprint"`
		} `json:"scope"`
	} `json:"meta"`
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Key    string `json:"key"`
			Expiry string `json:"expiry"`
			Status string `json:"status"`
		} `json:"attributes"`
	} `json:"data"`
}

// -------- machines

type createMachineRequest struct {
	Data machineData `json:"data"`
}

type machineData struct {
	ID            string               `json:"id,omitempty"`
	Type          string               `json:"type"`
	Attributes    machineAttributes    `json:"attributes"`
	Relationships machineRelationships `json:"relationships"`
}

type machineAttributes struct {
	Fingerprint string `json:"fingerprint"`
	Platform    string `json:"platform"`
	Name        string `json:"name"`
}

type machineRelationships struct {
	License licenseRelationship `json:"license"`
}

type machineResponse struct {
	Data machineData `json:"data"`
}

type machinesListResponse struct {
	Data []machineData `json:"data"`
}
