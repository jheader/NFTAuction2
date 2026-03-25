// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import {Script} from "forge-std/Script.sol";
import "../src/NFTAuction.sol";
import "forge-std/console.sol";
import "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

contract DeployNFTAuction is Script {

    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        vm.startBroadcast(deployerPrivateKey);
        // 1. 部署实现合约
        NFTAuction implementation = new NFTAuction();
        // 2. 部署代理合约  
        ERC1967Proxy proxy = new ERC1967Proxy(address(implementation), abi.encodeCall(NFTAuction.initialize, ()));
         // 3. 初始化代理合约（通过代理调用）
        NFTAuction auction = NFTAuction(address(proxy));

        vm.stopBroadcast();

        console.log("NFTAuction Implementation deployed to:", address(implementation));
        console.log("NFTAuction Proxy deployed to:", address(proxy));
    
    }

}