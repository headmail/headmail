import createClient from "openapi-fetch";
import type { paths } from "../generated/api-types";

const { GET, POST, PUT, DELETE } = createClient<paths>({ baseUrl: "/api" });

// Campaigns
export const getCampaigns = async (params: paths["/campaigns"]["get"]["parameters"]["query"]) => {
    const resp = await GET("/campaigns", { params: { query: params } });
    return resp.data;
};
export const getCampaign = async (campaignID: string) => {
    const resp = await GET("/campaigns/{campaignID}", { params: { path: { campaignID } } });
    return resp.data;
};

// Lists
export const getLists = async (params: paths["/lists"]["get"]["parameters"]["query"]) => {
    const resp = await GET("/lists", { params: { query: params } });
    return resp.data;
};
export const createList = async (req: paths["/lists"]["post"]["parameters"]["body"]["list"]) => {
    const resp = await POST("/lists", { params: { body: { list: req } } });
    return resp.data;
};
export const getList = async (listID: string) => {
    const resp = await GET("/lists/{listID}", { params: { path: { listID } } });
    return resp.data;
};
export const updateList = async (listID: string, req: paths["/lists/{listID}"]["put"]["parameters"]["body"]["list"]) => {
    const resp = await PUT("/lists/{listID}", { params: { path: { listID } , body: { list: req } } });
    return resp.data;
};
export const deleteList = async (listID: string) => {
    const resp = await DELETE("/lists/{listID}", { params: { path: { listID } } });
    return resp.data;
};

// Subscribers
export const getSubscribers = async (params: paths["/subscribers"]["get"]["parameters"]["query"]) => {
    const resp = await GET("/subscribers", { params: { query: params } });
    return resp.data;
};
export const getSubscriber = async (subscriberID: string) => {
    const resp = await GET("/subscribers/{subscriberID}", { params: { path: { subscriberID } } });
    return resp.data;
};
export const updateSubscriber = async (subscriberID: string, req: paths["/subscribers/{subscriberID}"]["put"]["parameters"]["body"]["subscriber"]) => {
    const resp = await PUT("/subscribers/{subscriberID}", { params: { path: { subscriberID }, body: { subscriber: req } } });
    return resp.data;
};
export const deleteSubscriber = async (subscriberID: string) => {
    const resp = await DELETE("/subscribers/{subscriberID}", { params: { path: { subscriberID } } });
    return resp.data;
};

// Templates
export const getTemplates = async (params: paths["/templates"]["get"]["parameters"]["query"]) => {
    const resp = await GET("/templates", { params: { query: params } });
    return resp.data;
};
export const createTemplate = async (req: paths["/templates"]["post"]["parameters"]["body"]["template"]) => {
    const resp = await POST("/templates", { params: { body: { template: req } } });
    return resp.data;
};
export const getTemplate = async (templateID: string) => {
    const resp = await GET("/templates/{templateID}", { params: { path: { templateID } } });
    return resp.data;
};
export const updateTemplate = async (templateID: string, req: paths["/templates/{templateID}"]["put"]["parameters"]["body"]["template"]) => {
    const resp = await PUT("/templates/{templateID}", { params: { path: { templateID }, body: { template: req } } });
    return resp.data;
};
export const deleteTemplate = async (templateID: string) => {
    const resp = await DELETE("/templates/{templateID}", { params: { path: { templateID } } });
    return resp.data;
};
