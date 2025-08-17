/**
 * Copyright 2025 JC-Lab
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

import createClient from "openapi-fetch";
import type {paths} from "../generated/api-types";

const {
    GET,
    POST,
    PUT,
    DELETE,
    PATCH,
} = createClient<paths>({baseUrl: "/api"});

// Campaigns
export const getCampaigns = async (params: paths["/campaigns"]["get"]["parameters"]["query"]) => {
    const resp = await GET("/campaigns", {params: {query: params}});
    return resp.data;
};
export const getCampaign = async (campaignID: string) => {
    const resp = await GET("/campaigns/{campaignID}", {params: {path: {campaignID}}});
    return resp.data;
};

// Get stats for a single campaign
export const getCampaignStats = async (
    campaignID: string,
    params?: paths["/campaigns/{campaignID}/stats"]["get"]["parameters"]["query"]
) => {
    const resp = await GET("/campaigns/{campaignID}/stats", {
        params: {
            path: { campaignID },
            query: params,
        },
    });
    return resp.data;
};
export const createCampaign = async (req: paths["/campaigns"]["post"]["requestBody"]["content"]["application/json"]) => {
    const resp = await POST("/campaigns", {
        body: req,
        headers: {
            'Content-Type': 'application/json',
        }
    });
    return resp.data;
};
export const updateCampaign = async (campaignID: string, req: paths["/campaigns/{campaignID}"]["put"]["requestBody"]["content"]["application/json"]) => {
    const resp = await PUT("/campaigns/{campaignID}", {
        params: {
            path: {campaignID},
        },
        body: req,
        headers: {
            'Content-Type': 'application/json',
        }
    });
    return resp.data;
};
export const deleteCampaign = async (campaignID: string) => {
    const resp = await DELETE("/campaigns/{campaignID}", {params: {path: {campaignID}}});
    return resp.data;
};

export const createCampaignDeliveries = async (
    campaignID: string,
    req: paths["/campaigns/{campaignID}/deliveries"]["post"]["requestBody"]["content"]["application/json"]
) => {
    const resp = await POST("/campaigns/{campaignID}/deliveries", {
        params: { path: { campaignID } },
        body: req,
        headers: {
            "Content-Type": "application/json",
        },
    });
    return resp.data;
};

 // Get deliveries for a campaign
 export const getCampaignDeliveries = async (
     campaignID: string,
     params?: paths["/campaigns/{campaignID}/deliveries"]["get"]["parameters"]["query"]
 ) => {
     const resp = await GET("/campaigns/{campaignID}/deliveries", {
         params: {
             path: { campaignID },
             query: params,
         },
     });
     return resp.data;
 };
 
 // Send a delivery immediately (synchronous)
 export const sendDeliveryNow = async (deliveryID: string) => {
     const resp = await fetch(`/api/deliveries/${deliveryID}/send-now`, {
         method: "POST",
         headers: { "Content-Type": "application/json" },
     });
     if (!resp.ok) {
         const text = await resp.text();
         throw new Error(text || `send-now failed: ${resp.status}`);
     }
     return resp.json();
 };
 
 // Retry a delivery immediately (synchronous)
 export const retryDelivery = async (deliveryID: string) => {
     const resp = await fetch(`/api/deliveries/${deliveryID}/retry`, {
         method: "POST",
         headers: { "Content-Type": "application/json" },
     });
     if (!resp.ok) {
         const text = await resp.text();
         throw new Error(text || `retry failed: ${resp.status}`);
     }
     return resp.json();
 };

// Lists
export const getLists = async (params: paths["/lists"]["get"]["parameters"]["query"]) => {
    const resp = await GET("/lists", {params: {query: params}});
    return resp.data;
};
export const createList = async (req: paths["/lists"]["post"]["requestBody"]["content"]["application/json"]) => {
    const resp = await POST("/lists", {
        body: req,
        headers: {
            'Content-Type': 'application/json',
        }
    });
    return resp.data;
};
export const getList = async (listID: string) => {
    const resp = await GET("/lists/{listID}", {params: {path: {listID}}});
    return resp.data;
};
export const updateList = async (listID: string, req: paths["/lists/{listID}"]["put"]["requestBody"]["content"]["application/json"]) => {
    const resp = await PUT("/lists/{listID}", {
        params: {
            path: {listID},
        },
        body: req,
        headers: {
            'Content-Type': 'application/json',
        }
    });
    return resp.data;
};
export const deleteList = async (listID: string) => {
    const resp = await DELETE("/lists/{listID}", {params: {path: {listID}}});
    return resp.data;
};

// Subscribers
export const getSubscribers = async (params: paths["/subscribers"]["get"]["parameters"]["query"]) => {
    const resp = await GET("/subscribers", {params: {query: params}});
    return resp.data;
};
export const getSubscriber = async (subscriberID: string) => {
    const resp = await GET("/subscribers/{subscriberID}", {params: {path: {subscriberID}}});
    return resp.data;
};
export const createSubscribers = async (req: paths["/subscribers"]["post"]["requestBody"]["content"]["application/json"]) => {
    const resp = await POST("/subscribers", {
        body: req,
        headers: {
            'Content-Type': 'application/json',
        }
    });
    return resp.data;
};
export const updateSubscriber = async (subscriberID: string, req: paths["/subscribers/{subscriberID}"]["put"]["requestBody"]["content"]["application/json"]) => {
    const resp = await PUT("/subscribers/{subscriberID}", {
        params: {
            path: {subscriberID},
        },
        body: req,
        headers: {
            'Content-Type': 'application/json',
        }
    });
    return resp.data;
};
export const deleteSubscriber = async (subscriberID: string) => {
    const resp = await DELETE("/subscribers/{subscriberID}", {params: {path: {subscriberID}}});
    return resp.data;
};

// Templates
export const getTemplates = async (params: paths["/templates"]["get"]["parameters"]["query"]) => {
    const resp = await GET("/templates", {params: {query: params}});
    return resp.data;
};
export const createTemplate = async (req: paths["/templates"]['post']['requestBody']['content']['application/json']) => {
    const resp = await POST("/templates", {
        body: req,
        headers: {
            'Content-Type': 'application/json',
        }
    });
    return resp.data;
};
export const getTemplate = async (templateID: string) => {
    const resp = await GET("/templates/{templateID}", {params: {path: {templateID}}});
    return resp.data;
};
export const updateTemplate = async (templateID: string, req: paths["/templates/{templateID}"]['put']['requestBody']['content']['application/json']) => {
    const resp = await PUT("/templates/{templateID}", {
        params: {
            path: {templateID},
        },
        body: req,
        headers: {
            'Content-Type': 'application/json',
        }
    });
    return resp.data;
};
export const deleteTemplate = async (templateID: string) => {
    const resp = await DELETE("/templates/{templateID}", {params: {path: {templateID}}});
    return resp.data;
};

// Preview template server-side (renders template with sample data).
// Uses generated types from ../generated/api-types to avoid `any`.
export const previewTemplate = async (
    req: paths["/templates/preview"]["post"]["requestBody"]["content"]["application/json"]
): Promise<
    paths["/templates/preview"]["post"]["responses"][200]["content"]["application/json"]
> => {
    const resp = await POST("/templates/preview", {
        body: req,
        headers: {
            "Content-Type": "application/json",
        },
    });
    return resp.data!!;
};

// Subscribers of a specific list
export const getSubscribersOfList = async (listID: string, params?: paths["/lists/{listID}/subscribers"]["get"]["parameters"]["query"]) => {
    const resp = await GET("/lists/{listID}/subscribers", { params: { path: { listID }, query: params }});
    return resp.data;
};

// Patch subscribers in a list (add/remove)
export const patchListSubscribers = async (listID: string, req: paths["/lists/{listID}/subscribers"]["patch"]["requestBody"]["content"]["application/json"]) => {
    const resp = await PATCH("/lists/{listID}/subscribers", {
        params: { path: { listID } },
        body: req,
        headers: { "Content-Type": "application/json" },
    });
    return resp.data;
};

// Replace subscribers in a list (atomic)
export const replaceListSubscribers = async (listID: string, req: paths["/lists/{listID}/subscribers"]["put"]["requestBody"]["content"]["application/json"]) => {
    const resp = await (createClient<paths>({ baseUrl: "/api" })).PUT("/lists/{listID}/subscribers", {
        params: { path: { listID } },
        body: req,
        headers: { "Content-Type": "application/json" },
    });
    return resp.data;
};
