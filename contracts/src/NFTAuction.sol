// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;
import "../lib/openzeppelin-contracts-upgradeable/contracts/token/ERC721/ERC721Upgradeable.sol";
import "../lib/openzeppelin-contracts-upgradeable/contracts/proxy/utils/Initializable.sol";
import "../lib/openzeppelin-contracts-upgradeable/contracts/access/OwnableUpgradeable.sol";
import "../lib/openzeppelin-contracts-upgradeable/contracts/proxy/utils/UUPSUpgradeable.sol";

contract  NFTAuction is Initializable,OwnableUpgradeable,UUPSUpgradeable{

    enum AuctionStatus {
         PENDING, // 拍卖待开始
         ACTIVE,  // 拍卖进行中
         ENDED,   // 拍卖已结束
         SOLD,    // 拍卖成功，NFT已售出
         UNSOLD,  // 拍卖结束但未售出
         CANCELLED // 拍卖取消
    }

    struct Auction {
        address nftContract; // NFT合约地址
        uint256 tokenId;     // NFT TokenID
        address seller;      // 卖家地址
        uint256 startTime;   // 开始时间（时间戳）
        uint256 duration;    // 拍卖时长（秒）
        uint256 startPrice;  // 起拍价（wei）
        uint256 highestBid;  // 当前最高价
        address highestBidder; // 当前最高出价者
        AuctionStatus status;  // 拍卖状态
    }

    // NFT(合约+TokenID) -> 拍卖ID
     mapping(bytes32 => uint256) public nftToken2AuctionId;
     // 全局拍卖ID计数器
    uint256 public auctionIdCounter;
    // 拍卖ID -> 拍卖数据
    mapping(uint256 => Auction) public auctionData;

     // 事件定义
    event AuctionCreated(uint256 indexed auctionId, address indexed nftContract, uint256 indexed tokenId, address seller, uint256 startPrice, uint256 startTime, uint256 duration);
    event BidPlaced(uint256 indexed auctionId, address indexed bidder, uint256 amount);
    event AuctionEnded(uint256 indexed auctionId, address indexed winner, uint256 finalPrice);
    event AuctionCancelled(uint256 indexed auctionId);

   // 初始化函数（替代构造函数）
    function initialize() public initializer {
        __Ownable_init(msg.sender);
        __UUPSUpgradeable_init();
        auctionIdCounter = 1; // 拍卖ID从1开始

    }

     // UUPS升级授权（仅所有者可升级）
    function _authorizeUpgrade(address newImplementation) internal override onlyOwner {}

    // =============== 拍卖核心功能 ===============
    // 创建拍卖逻辑 
    function createAuction(address nftcontract,uint256 tokenId, uint256 startTime, uint256 duration, uint256 startPrice) external {
        
        require(nftcontract != address(0), "Invalid NFT contract");
        require(tokenId !=0, "Invalid tokenId");
        require(startPrice>0,"Start price must be > 0");
        require(startTime > block.timestamp, "Start time must be in future");
        require(duration > 0, "Duration must be > 0");

        IERC721 nft = IERC721(nftcontract);
        require(nft.ownerOf(tokenId) == msg.sender, "Only NFT owner can create auction");
        //授权
        require(nft.isApprovedForAll(msg.sender, address(this)) || nft.getApproved(tokenId) == address(this), "Contract not approved");
         // 检查该NFT是否已有未结束的拍卖
        bytes32 auctionKey = keccak256(abi.encodePacked(nftcontract, tokenId));
        uint256 existingAuctionId = nftToken2AuctionId[auctionKey];
          if (existingAuctionId > 0) {
            Auction memory existingAuction = auctionData[existingAuctionId];
            require(existingAuction.status == AuctionStatus.ENDED || existingAuction.status == AuctionStatus.CANCELLED, "NFT already in auction");
        }
        //锁定到合约
        nft.transferFrom(msg.sender, address(this), tokenId);
        
        //生产新拍ID
        uint256 auctionId = auctionIdCounter++;
        // 创建拍卖数据
        Auction memory newAuction = Auction({
            nftContract: nftcontract,
            tokenId: tokenId,
            seller: msg.sender,
            startTime: startTime,               
            duration: duration,
            startPrice: startPrice,
            highestBid: 0,
            highestBidder: address(0),
            status: AuctionStatus.PENDING
        });
        nftToken2AuctionId[auctionKey] = auctionId;
        auctionData[auctionId] = newAuction;
        emit AuctionCreated(auctionIdCounter, nftcontract, tokenId, msg.sender, startPrice, startTime, duration);
        auctionIdCounter++;
    }


    //参与出价
    function bidAuction(uint256 auctionId) external payable {
        
        Auction storage auction = auctionData[auctionId];
        require(auction.seller != address(0),"Auction not exists");
        require(block.timestamp > auction.startTime,"Auction not started");
        require(block.timestamp <= auction.startTime + auction.duration, "Auction ended");
        require(auction.status == AuctionStatus.ACTIVE || auction.status == AuctionStatus.PENDING, "Auction invalid");
        require(msg.sender != auction.seller,"Seller cannot bid");
        require(msg.value > auction.highestBid, "Bid too low");

         // 初始化拍卖状态（首次出价）
        if (auction.status == AuctionStatus.PENDING) {
            auction.status = AuctionStatus.ACTIVE;
        }
        require(msg.value> auction.highestBid,"need than last value high");
        //回退上出价高者
        if(auction.highestBidder != address(0)){

            payable(auction.highestBidder).transfer(auction.highestBid);
        }
         // 更新最高出价
        auction.highestBid = msg.value;
        auction.highestBidder = msg.sender;
        emit BidPlaced(auctionId, msg.sender, msg.value);
    }

    //结束拍卖
    function endAuction(uint256 auctionId) external {

        Auction storage auction = auctionData[auctionId];
        require(auction.seller != address(0), "Auction not exists");
        require(block.timestamp > auction.startTime + auction.duration, "Auction not ended");
        require(auction.status == AuctionStatus.ACTIVE, "Auction not active");
        auction.status = AuctionStatus.ENDED;

        if (auction.highestBidder != address(0)) {
             // 1. 卖家获得资金
            payable(auction.seller).transfer(auction.highestBid);
            IERC721(auction.nftContract).transferFrom(address(this), auction.highestBidder, auction.tokenId);
            emit AuctionEnded(auctionId, auction.highestBidder, auction.highestBid);

        } else {

             // 无出价者，卖家取回NFT
            IERC721(auction.nftContract).transferFrom(address(this), auction.seller, auction.tokenId);
            emit AuctionEnded(auctionId, address(0), 0);
        }

    }


    //取消拍卖（仅开始前可操作）

    function cancelAuction(uint256 auctionId) external {
        Auction storage auction = auctionData[auctionId];
        require(auction.seller != address(0), "Auction not exists");
        require(msg.sender == auction.seller, "Only seller can cancel");
        require(block.timestamp < auction.startTime, "Auction already started");
        require(auction.status == AuctionStatus.PENDING, "Auction already active");
        auction.status = AuctionStatus.CANCELLED;
        // 返还NFT给卖家
        IERC721(auction.nftContract).transferFrom(address(this), auction.seller, auction.tokenId);

        emit AuctionCancelled(auctionId);
    }
}