// SPDX-License-Identifier: UNLICENSED

pragma solidity ^0.8.0;
import "@openzeppelin/contracts/token/ERC1155/ERC1155.sol";

contract ProcessingFactory is ERC1155 {
    address public contractDeployer; // 合约部署者地址

    uint256 public maxGoodsId; // 当前已经铸币的最大的币 id

    /**
     * 币的属性, 使用Goods 结构体整合
     * goodsID 币的id
     * amount 发行数量
     * publishTime 发行时间, 时间戳
     * publisher 发行方
     * name 通证名称
     * labelID 通证标签ID， 展示用的通证ID
     * goodsType 通证类型  0:FT(ERC20) 1:NFT(ERC721)
     * remark 币的备注
     */
    struct Goods {
        uint256 goodsID;
        uint256 amount;
        int32 goodsType;
        int64 publishTime;
        string publisher;
        string name;
        string labelID;
        string remark;
    }

    /**
     * 余额，地址对应的合约里面币的余额
     * account 用户地址
     * goodsID 币的id
     * balance 余额数量
     */
    struct AccountBalance {
        address account;
        uint256 goodsID;
        uint256 balance;
    }

    // id已经被使用的记录
    mapping(uint256 => bool) goodsIdUsed;

    // 发币白名单地址
    mapping(address => bool) mintWhiteList;

    // 通证（币）属性
    mapping(uint256 => Goods) attribute;

    // 币 对应的当前hash
    mapping(uint256 => string[]) public goodsHash;

    // 币 对应的来源hash，如果币是由一些溯源来源转换而来，该值记录币的转换之前的来源
    mapping(uint256 => string[]) public sourceHash;

    // 发行币成功后，调用事件通知
    event addNewGoodsResult(
        address owner,
        uint256 goodsID,
        uint256 amount,
        int32 goodsType,
        int64 publishTime,
        string[] hash,
        string[] source,
        string publisher,
        string name,
        string labelID,
        string remark
    );

    event batchAddNewGoodsResult(
        address owner,
        uint256[] goodsIDs,
        uint256[] amounts,
        int32 goodsType,
        int64 publishTime,
        string[] hash,
        string[] source,
        string publisher,
        string[] names,
        string[] labelIDs,
        string[] remarks
    );

    // 用户余额变动后，通过event发出事件通知，后续服务记录解析
    event balanceResult(AccountBalance[] balanceList);

    // 交易成功后，调用事件通知
    event batchTransferResult(
        address from,
        address to,
        uint256[] ids,
        uint256[] amounts,
        bytes data
    );

    constructor() ERC1155("") {
        contractDeployer = _msgSender();
        mintWhiteList[contractDeployer] = true;
    }

    /**
     * 铸币, 传输完整参数
     * owner 币的属于人
     * goodsID 币的id
     * amount 币的数量，nft=1，ft为实际数量； 给 owner铸造id为goodsID的币，数量为amount, 如果amount=1，为nft，如果大于1，为ft
     * goodsType 通证类型  0:FT(ERC20) 1:NFT(ERC721)
     * publishTime 发行时间, 时间戳
     * hash 币对应的hash
     * source 币对应的来源hash
     * publisher 发行方
     * name 通证名称
     * labelID 通证标签ID
     * remark 币的备注
     */
    function addNewGoodsAssignGoodsID(
        address owner,
        uint256 goodsID,
        uint256 amount,
        int32 goodsType,
        int64 publishTime,
        string[] memory hash,
        string[] memory source,
        string memory publisher,
        string memory name,
        string memory labelID,
        string memory remark
    ) public returns (uint256) {
        require(
            isInMintWhiteList(_msgSender()),
            "ProcessingFactory: msgSender not in mintWhiteList"
        );
        require(
            (false == goodsIdUsed[goodsID]),
            "ProcessingFactory: goodsID is already been used"
        );
        goodsIdUsed[goodsID] = true;

        if (maxGoodsId < goodsID) {
            maxGoodsId = goodsID;
        }

        attribute[goodsID] = Goods(
            goodsID,
            amount,
            goodsType,
            publishTime,
            publisher,
            name,
            labelID,
            remark
        );

        goodsHash[goodsID] = hash;
        sourceHash[goodsID] = source;
        _mint(owner, goodsID, amount, "");

        AccountBalance[] memory balanceList = new AccountBalance[](1);
        balanceList[0] = AccountBalance(owner, goodsID, amount);

        emit addNewGoodsResult(
            owner,
            goodsID,
            amount,
            goodsType,
            publishTime,
            hash,
            source,
            publisher,
            name,
            labelID,
            remark
        );

        emit balanceResult(balanceList);
        return goodsID;
    }

    /**
     * 批量发币
     * owner 为本次将币发到哪个地址的账户
     * goodsIDs为数组，数组内每个元素都是一个币id，每个币id只能使用一次，不能重复发币，如果重复将会返回一个error
     * amounts, names, labelIDs, remarks 都为数组，和goodsIDs一一对应，
     *  例如: goodsIDs[0]对应的数量为amounts[0],对应的币名称为names[0]，对应的通证标签id为labelIDs[0]，对应的备注信息为remarks[0]
     * 这一批goodsIDs对应的币，每一个币对应的goodsType，publishTime，hash，source，publisher都是相同值
     * goodsType 为通证类型  0:FT(ERC20) 1:NFT(ERC721)，publishTime 发行时间, 时间戳
     * hash 币对应的存证溯源hash，source 币对应的合并存证的来源hash，publisher 为发行方名称
     *
     * 发币成功后，发出两个event, batchAddNewGoodsResult和balanceResult
     */
    function batchAddNewGoodsAssignGoodsID(
        address owner,
        uint256[] memory goodsIDs,
        uint256[] memory amounts,
        int32 goodsType,
        int64 publishTime,
        string[] memory hash,
        string[] memory source,
        string memory publisher,
        string[] memory names,
        string[] memory labelIDs,
        string[] memory remarks
    ) public {
        require(
            isInMintWhiteList(_msgSender()),
            "ProcessingFactory: msgSender not in mintWhiteList"
        );

        if (maxGoodsId < goodsIDs[goodsIDs.length - 1]) {
            maxGoodsId = goodsIDs[goodsIDs.length - 1];
        }

        AccountBalance[] memory balanceList = new AccountBalance[](
            goodsIDs.length
        );
        for (uint256 i = 0; i < goodsIDs.length; ++i) {
            uint256 goodsID = goodsIDs[i];
            require(
                (false == goodsIdUsed[goodsID]),
                "ProcessingFactory: goodsID is already been used"
            );
            goodsIdUsed[goodsID] = true;

            attribute[goodsID] = Goods(
                goodsID,
                amounts[i],
                goodsType,
                publishTime,
                publisher,
                names[i],
                labelIDs[i],
                remarks[i]
            );

            goodsHash[goodsID] = hash;
            sourceHash[goodsID] = source;
            balanceList[i] = AccountBalance(owner, goodsID, amounts[i]);
        }
        _mintBatch(owner, goodsIDs, amounts, "");

        emit batchAddNewGoodsResult(
            owner,
            goodsIDs,
            amounts,
            goodsType,
            publishTime,
            hash,
            source,
            publisher,
            names,
            labelIDs,
            remarks
        );

        emit balanceResult(balanceList);
    }

    /**
     * 铸币，使用自增的goodsId, 推荐使用
     * owner 币的属于人
     * amount 币的数量，nft=1，ft为实际数量
     * goodsType 通证类型  0:FT(ERC20) 1:NFT(ERC721)
     * publishTime 发行时间, 时间戳
     * hash 币对应的hash
     * source 币对应的来源hash
     * publisher 发行方
     * name 通证名称
     * labelID 通证标签ID
     * remark 币的备注
     *
     * return
     * uint256 通证id
     */
    function addNewGoods(
        address owner,
        uint256 amount,
        int32 goodsType,
        int64 publishTime,
        string[] memory hash,
        string[] memory source,
        string memory publisher,
        string memory name,
        string memory labelID,
        string memory remark
    ) public returns (uint256) {
        uint256 goodsID = maxGoodsId + 1;
        return
            addNewGoodsAssignGoodsID(
                owner,
                goodsID,
                amount,
                goodsType,
                publishTime,
                hash,
                source,
                publisher,
                name,
                labelID,
                remark
            );
    }

    /**
     * 交易之后调用event事件，便于chain33 日志解析
     * from 从哪个账户扣款
     * to 给哪个账户增加余额
     * ids 操作币的id
     * amounts 每个币交易的量，位置一一对应
     * data 交易备注 空备注可以为"0x"
     */
    function BatchTransferWithEvent(
        address from,
        address to,
        uint256[] memory ids,
        uint256[] memory amounts,
        bytes memory data
    ) public {
        safeBatchTransferFrom(from, to, ids, amounts, data);

        AccountBalance[] memory balanceList = new AccountBalance[](
            2 * ids.length
        );
        for (uint256 i = 0; i < ids.length; ++i) {
            uint256 fromBalance = balanceOf(from, ids[i]);
            balanceList[2 * i] = AccountBalance(from, ids[i], fromBalance);

            uint256 toBalance = balanceOf(to, ids[i]);
            balanceList[2 * i + 1] = AccountBalance(to, ids[i], toBalance);
        }

        emit batchTransferResult(from, to, ids, amounts, data);
        emit balanceResult(balanceList);
    }

    /**
     * 修改发币白名单地址
     * operator 白名单内操作人地址
     * approved 授权状态，true为已授权，false为移除授权
     */
    function setMintWhiteList(address operator, bool approved) public {
        // 判断调用者是合约部署者，才能修改白名单
        require(
            _msgSender() == contractDeployer,
            "ProcessingFactory: setMintWhiteList must only be called by the contractDeployer"
        );
        // 修改白名单内地址的授权状态
        mintWhiteList[operator] = approved;
    }

    /**
     * 查询某个地址是否白名单地址
     * operator 白名单内操作人地址
     * approved 授权状态，true为已授权，false为移除授权
     */
    function isInMintWhiteList(address operator) public view returns (bool) {
        // 查询地址的授权状态
        return mintWhiteList[operator];
    }

    // 查询币对应的hash
    function getGoodsHash(uint256 goodsID)
        public
        view
        returns (string[] memory)
    {
        return goodsHash[goodsID];
    }

    // 查询币对应的来源hash
    function getsourceHash(uint256 goodsID)
        public
        view
        returns (string[] memory)
    {
        return sourceHash[goodsID];
    }

    // 查询币对应的属性
    function getGoodsAttribute(uint256 goodsID)
        public
        view
        returns (Goods memory)
    {
        return attribute[goodsID];
    }

    // 查询当前已经发行了的最大的币ID
    function getMaxGoodsID() public view returns (uint256) {
        return maxGoodsId;
    }
}
