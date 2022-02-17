#! /usr/bin/env node

import toJsonSchema from "@openapi-contrib/openapi-schema-to-json-schema";

let schema = {
    "properties": {
        "apiVersion": {
            "description": "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources",
            "type": "string"
        },
        "kind": {
            "description": "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds",
            "type": "string"
        },
        "metadata": {
            "type": "object"
        },
        "spec": {
            "description": "Spec describes the suppply chain. More info: https://cartographer.sh/docs/latest/reference/workload/#clustersupplychain",
            "properties": {
                "params": {
                    "description": "Additional parameters. See: https://cartographer.sh/docs/latest/architecture/#parameter-hierarchy",
                    "items": {
                        "properties": {
                            "default": {
                                "description": "DefaultValue of the parameter. Causes the parameter to be optional; If the Owner does not specify this parameter, this value is used.",
                                "x-kubernetes-preserve-unknown-fields": true
                            },
                            "name": {
                                "description": "Name of the parameter. Should match a template parameter name.",
                                "type": "string"
                            },
                            "value": {
                                "description": "Value of the parameter. If specified, owner properties are ignored.",
                                "x-kubernetes-preserve-unknown-fields": true
                            }
                        },
                        "required": [
                            "name"
                        ],
                        "type": "object"
                    },
                    "type": "array"
                },
                "resources": {
                    "description": "Resources that are responsible for bringing the application to a deliverable state.",
                    "items": {
                        "properties": {
                            "configs": {
                                "description": "Configs is a list of references to other 'config' resources in this list. A config resource has the kind ClusterConfigTemplate \n In a template, configs can be consumed as: $(configs.<name>.config)$ \n If there is only one image, it can be consumed as: $(config)$",
                                "items": {
                                    "properties": {
                                        "name": {
                                            "type": "string"
                                        },
                                        "resource": {
                                            "type": "string"
                                        }
                                    },
                                    "required": [
                                        "name",
                                        "resource"
                                    ],
                                    "type": "object"
                                },
                                "type": "array"
                            },
                            "images": {
                                "description": "Images is a list of references to other 'image' resources in this list. An image resource has the kind ClusterImageTemplate \n In a template, images can be consumed as: $(images.<name>.image)$ \n If there is only one image, it can be consumed as: $(image)$",
                                "items": {
                                    "properties": {
                                        "name": {
                                            "type": "string"
                                        },
                                        "resource": {
                                            "type": "string"
                                        }
                                    },
                                    "required": [
                                        "name",
                                        "resource"
                                    ],
                                    "type": "object"
                                },
                                "type": "array"
                            },
                            "name": {
                                "description": "Name of the resource. Used as a reference for inputs, as well as being the name presented in workload statuses to identify this resource.",
                                "type": "string"
                            },
                            "params": {
                                "description": "Params are a list of parameters to provide to the template in TemplateRef Template params do not have to be specified here, unless you want to force a particular value, or add a default value. \n Parameters are consumed in a template with the syntax: $(params.<name>)$",
                                "items": {
                                    "properties": {
                                        "default": {
                                            "description": "DefaultValue of the parameter. Causes the parameter to be optional; If the Owner does not specify this parameter, this value is used.",
                                            "x-kubernetes-preserve-unknown-fields": true
                                        },
                                        "name": {
                                            "description": "Name of the parameter. Should match a template parameter name.",
                                            "type": "string"
                                        },
                                        "value": {
                                            "description": "Value of the parameter. If specified, owner properties are ignored.",
                                            "x-kubernetes-preserve-unknown-fields": true
                                        }
                                    },
                                    "required": [
                                        "name"
                                    ],
                                    "type": "object"
                                },
                                "type": "array"
                            },
                            "sources": {
                                "description": "Sources is a list of references to other 'source' resources in this list. A source resource has the kind ClusterSourceTemplate \n In a template, sources can be consumed as: $(sources.<name>.url)$ and $(sources.<name>.revision)$ \n If there is only one source, it can be consumed as: $(source.url)$ and $(source.revision)$",
                                "items": {
                                    "properties": {
                                        "name": {
                                            "type": "string"
                                        },
                                        "resource": {
                                            "type": "string"
                                        }
                                    },
                                    "required": [
                                        "name",
                                        "resource"
                                    ],
                                    "type": "object"
                                },
                                "type": "array"
                            },
                            "templateRef": {
                                "description": "TemplateRef identifies the template used to produce this resource",
                                "properties": {
                                    "kind": {
                                        "description": "Kind of the template to apply",
                                        "enum": [
                                            "ClusterSourceTemplate",
                                            "ClusterImageTemplate",
                                            "ClusterTemplate",
                                            "ClusterConfigTemplate"
                                        ],
                                        "type": "string"
                                    },
                                    "name": {
                                        "description": "Name of the template to apply",
                                        "minLength": 1,
                                        "type": "string"
                                    }
                                },
                                "required": [
                                    "kind",
                                    "name"
                                ],
                                "type": "object"
                            }
                        },
                        "required": [
                            "name",
                            "templateRef"
                        ],
                        "type": "object"
                    },
                    "type": "array"
                },
                "selector": {
                    "additionalProperties": {
                        "type": "string"
                    },
                    "description": "Specifies the label key-value pairs used to select workloads See: https://cartographer.sh/docs/v0.1.0/architecture/#selectors",
                    "type": "object"
                },
                "serviceAccountRef": {
                    "description": "ServiceAccountName refers to the Service account with permissions to create resources submitted by the supply chain. \n If not set, Cartographer will use serviceAccountName from supply chain. \n If that is also not set, Cartographer will use the default service account in the workload's namespace.",
                    "properties": {
                        "name": {
                            "description": "Name of the service account being referred to",
                            "type": "string"
                        },
                        "namespace": {
                            "description": "Namespace of the service account being referred to if omitted, the Owner's namespace is used.",
                            "type": "string"
                        }
                    },
                    "required": [
                        "name"
                    ],
                    "type": "object"
                }
            },
            "required": [
                "resources",
                "selector"
            ],
            "type": "object"
        },
        "status": {
            "description": "Status conforms to the Kubernetes conventions: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties",
            "properties": {
                "conditions": {
                    "items": {
                        "description": "Condition contains details for one aspect of the current state of this API Resource. --- This struct is intended for direct use as an array at the field path .status.conditions.  For example, type FooStatus struct{ // Represents the observations of a foo's current state. // Known .status.conditions.type are: \"Available\", \"Progressing\", and \"Degraded\" // +patchMergeKey=type // +patchStrategy=merge // +listType=map // +listMapKey=type Conditions []metav1.Condition `json:\"conditions,omitempty\" patchStrategy:\"merge\" patchMergeKey:\"type\" protobuf:\"bytes,1,rep,name=conditions\"` \n // other fields }",
                        "properties": {
                            "lastTransitionTime": {
                                "description": "lastTransitionTime is the last time the condition transitioned from one status to another. This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.",
                                "format": "date-time",
                                "type": "string"
                            },
                            "message": {
                                "description": "message is a human readable message indicating details about the transition. This may be an empty string.",
                                "maxLength": 32768,
                                "type": "string"
                            },
                            "observedGeneration": {
                                "description": "observedGeneration represents the .metadata.generation that the condition was set based upon. For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date with respect to the current state of the instance.",
                                "format": "int64",
                                "minimum": 0,
                                "type": "integer"
                            },
                            "reason": {
                                "description": "reason contains a programmatic identifier indicating the reason for the condition's last transition. Producers of specific condition types may define expected values and meanings for this field, and whether the values are considered a guaranteed API. The value should be a CamelCase string. This field may not be empty.",
                                "maxLength": 1024,
                                "minLength": 1,
                                "pattern": "^[A-Za-z]([A-Za-z0-9_,:]*[A-Za-z0-9_])?$",
                                "type": "string"
                            },
                            "status": {
                                "description": "status of the condition, one of True, False, Unknown.",
                                "enum": [
                                    "True",
                                    "False",
                                    "Unknown"
                                ],
                                "type": "string"
                            },
                            "type": {
                                "description": "type of condition in CamelCase or in foo.example.com/CamelCase. --- Many .condition.type values are consistent across resources like Available, but because arbitrary conditions can be useful (see .node.status.conditions), the ability to deconflict is important. The regex it matches is (dns1123SubdomainFmt/)?(qualifiedNameFmt)",
                                "maxLength": 316,
                                "pattern": "^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])$",
                                "type": "string"
                            }
                        },
                        "required": [
                            "lastTransitionTime",
                            "message",
                            "reason",
                            "status",
                            "type"
                        ],
                        "type": "object"
                    },
                    "type": "array"
                },
                "observedGeneration": {
                    "format": "int64",
                    "type": "integer"
                }
            },
            "type": "object"
        }
    },
    "required": [
        "metadata",
        "spec"
    ],
    "type": "object"
};

let convertedSchema = toJsonSchema(schema);

console.log(JSON.stringify(convertedSchema));