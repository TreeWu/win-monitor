// tslint:disable
/**
 * win-monitor
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * The version of the OpenAPI document: 1.0.0
 *
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */


import {MainMonitorConf} from './main-monitor-conf';

/**
 *
 * @export
 * @interface MainHostModel
 */
export interface MainHostModel {
    /**
     * 系统
     * @type {string}
     * @memberof MainHostModel
     */
    OS?: string;
    /**
     * 配置信息
     * @type {MainMonitorConf}
     * @memberof MainHostModel
     */
    config: MainMonitorConf;
    /**
     * 自定义主机名
     * @type {string}
     * @memberof MainHostModel
     */
    customName?: string;
    /**
     * 首次注册时间
     * @type {number}
     * @memberof MainHostModel
     */
    firstRegisterTime?: number;
    /**
     * 主机唯一标识
     * @type {string}
     * @memberof MainHostModel
     */
    hostID: string;
    /**
     * 主机名
     * @type {string}
     * @memberof MainHostModel
     */
    hostname?: string;
    /**
     *
     * @type {number}
     * @memberof MainHostModel
     */
    id?: number;
    /**
     * 是否推送告警
     * @type {boolean}
     * @memberof MainHostModel
     */
    notifyPush?: boolean;
    /**
     * 系统平台
     * @type {string}
     * @memberof MainHostModel
     */
    platform?: string;
    /**
     * 系统家族
     * @type {string}
     * @memberof MainHostModel
     */
    platformFamily?: string;
    /**
     * 系统版本
     * @type {string}
     * @memberof MainHostModel
     */
    platformVersion?: string;
}


