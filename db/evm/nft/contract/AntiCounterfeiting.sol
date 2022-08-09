pragma solidity ^0.8.0;
import "../@openzeppelin/contracts/token/ERC1155/ERC1155.sol"; // 如果在本地构建，使用这个地址

// 功能清单
//  批量发币
//  批量转账
//  查询币的属性
//  通过验证码转账
//  设置批次号与验证码的对应关系
//  查询批次内验证码是否被使用
//  设置发币白名单地址
//  查询一个地址是否被设置为白名单地址
//
// 已实现功能，erc1155已经实现，可直接调用
//  设置一个地址A为另一个地址B的授权人
//  查询一个地址A是否为另一个地址B的授权人

contract AntiCounterfeiting is ERC1155 {
    uint256 public maxGoodsId; // 当前已经铸币的最大的币 id
    address public contractDeployer; // 合约部署者地址

    /**
     * 币的属性, 使用Goods 结构体整合
     * operator 操作人
     * publisherAddress 币的初始是由哪个地址发行的
     * goodsID 币的id
     * amount 发行数量
     * goodsType 通证类型  0:FT(ERC20) 1:NFT(ERC721)
     * name 通证名称
     * labelID 通证标签ID， 展示用的通证ID
     * batchNumber 批次号
     * image 通证对应的图片url
     * isUsed 该防伪币的防伪属性是否被使用
     */
    struct Goods {
        address operator;
        address publisherAddress;
        uint256 goodsID;
        uint256 amount;
        int32 goodsType;
        string name;
        string labelID;
        string batchNumber;
        string image;
        string traceHash;
        bool isUsed;
    }

    // 发币白名单地址
    mapping(address => bool) mintWhiteList;

    // 通证（币）属性
    mapping(uint256 => Goods) attribute;

    // 防伪通证的领取验证码, 一个批次对应一批验证码，并记录每个验证码的使用状态
    mapping(string => mapping(string => uint8)) batchVerificationCode;

    // 防伪通证的id与批次的对应关系
    mapping(uint256 => string) goodsBatchNumber;

    // 发行币成功后，调用事件通知
    event BatchMintResult(Goods[] mintGoodsList);

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

    // 用户余额变动后，通过event发出事件通知，后续服务记录解析
    event BalanceResult(AccountBalance[] balanceList);

    // 交易成功后，调用事件通知
    event BatchTransferResult(
        address from,
        address to,
        uint256[] ids,
        uint256[] amounts,
        bytes data
    );

    // 通过验证码验证通过并转账，修改goodsID的isUsed为true
    event GoodsUsedResult(
        address from,
        address to,
        uint256 id,
        string code,
        bool isUsed
    );

    /**
     * 批次号与验证码的数据，用于返回
     * batchNumber 批次号
     * code 验证码
     * status 验证码状态，0:验证码与批次未关联，1:验证码未使用, 2:验证码已使用
     */
    struct BatchNumberVerificationCode {
        string batchNumber;
        string code;
        uint8 status;
    }
    // 验证码事件，增加验证码或者消费验证码，发出此事件，便于后续服务解析
    event VerificationCodeResult(
        BatchNumberVerificationCode[] batchNumberVerificationCodeList
    );

    constructor() public ERC1155("") {
        maxGoodsId = 0;
        contractDeployer = _msgSender();
        mintWhiteList[contractDeployer] = true;
    }

    /**
     * 批量发币, 给一个地址一次发多个币,
     * 其他参数均为数组，币id为goodsIDs[0]的币，发行数量amount为amounts[0], 币的名称是names[0]
     * 同理，币id为goodsIDs[1]的币，发行数量amount为amounts[1], 币的名称是names[1]，通过这种方式一一对应，所有参数数组的长度必须长度相等
     *
     * goodsIDs 币的id
     * amounts 币的数量，nft=1，ft为实际数量； 给 owner铸造id为goodsID的币，数量为amount, 如果amount=1，为nft，如果大于1，为ft
     * goodsTypes 通证类型  0:FT(ERC20) 1:NFT(ERC721)
     * names 通证名称
     * labelIDs 通证标签ID
     * batchNumbers 批次号
     * images 通证对应的图片url
     * traceHashs 币对应的hash
     */
    function batchMintWithEvent(
        address owner,
        uint256[] memory goodsIDs,
        uint256[] memory amounts,
        int32[] memory goodsTypes,
        string[] memory names,
        string[] memory labelIDs,
        string[] memory batchNumbers,
        string[] memory images,
        string[] memory traceHashs
    ) public {
        require(
            isInMintWhiteList(_msgSender()),
            "AntiCounterfeiting: msgSender not in mintWhiteList"
        );

        for (uint256 i = 0; i < goodsIDs.length; ++i) {
            uint256 id = goodsIDs[i];
            require(
                (id > maxGoodsId),
                "AntiCounterfeiting: goodsID must be incremented"
            );
            maxGoodsId = id;
        }
        require(
            goodsIDs.length == amounts.length &&
            goodsIDs.length == goodsTypes.length &&
            goodsIDs.length == names.length &&
            goodsIDs.length == labelIDs.length &&
            goodsIDs.length == images.length &&
            goodsIDs.length == traceHashs.length,
            "AntiCounterfeiting: parameters length must be equal"
        );

        _mintBatch(owner, goodsIDs, amounts, "");
        address opetator = _msgSender();
        for (uint256 i = 0; i < goodsIDs.length; ++i) {
            uint256 goodsID = goodsIDs[i];
            attribute[goodsID] = Goods(
                opetator,
                owner,
                goodsID,
                amounts[i],
                goodsTypes[i],
                names[i],
                labelIDs[i],
                batchNumbers[i],
                images[i],
                traceHashs[i],
                false
            );
            goodsBatchNumber[goodsID] = batchNumbers[i];
        }

        mintEvent(
            owner,
            goodsIDs,
            amounts,
            goodsTypes,
            names,
            labelIDs,
            batchNumbers,
            images,
            traceHashs
        );
    }

    function mintEvent(
        address owner,
        uint256[] memory goodsIDs,
        uint256[] memory amounts,
        int32[] memory goodsTypes,
        string[] memory names,
        string[] memory labelIDs,
        string[] memory batchNumbers,
        string[] memory images,
        string[] memory traceHashs
    ) internal {
        address operator = _msgSender();
        Goods[] memory mintGoodsList = new Goods[](goodsIDs.length);
        AccountBalance[] memory balanceList = new AccountBalance[](
            goodsIDs.length
        );

        for (uint256 i = 0; i < goodsIDs.length; ++i) {
            uint256 goodsID = goodsIDs[i];
            mintGoodsList[i] = Goods(
                operator,
                owner,
                goodsID,
                amounts[i],
                goodsTypes[i],
                names[i],
                labelIDs[i],
                batchNumbers[i],
                images[i],
                traceHashs[i],
                false
            );

            balanceList[i] = AccountBalance(owner, goodsID, amounts[i]);
        }

        emit BatchMintResult(mintGoodsList);
        emit BalanceResult(balanceList);
    }

    /**
     * 交易之后调用event事件，便于chain33 日志解析
     * from 从哪个账户扣款
     * to 给哪个账户增加余额
     * ids 操作币的id
     * amounts 每个币交易的量，位置一一对应
     * data 交易备注
     */
    function batchTransferWithEvent(
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

        emit BatchTransferResult(from, to, ids, amounts, data);
        emit BalanceResult(balanceList);
    }

    /**
     * 通过验证码转账
     * from 从哪个账户扣款
     * to 给哪个账户增加余额
     * id 操作币的id
     * amount 币交易的量
     * data 交易备注
     * code 验证码
     */
    function transferWithVerificationCode(
        address from,
        address to,
        uint256 id,
        uint256 amount,
        bytes memory data,
        string memory code
    ) public {
        // 判断验证码是否存在和使用状态，不存在或者已使用，就返回error
        string memory batchNumber = goodsBatchNumber[id];
        require(
            1 == getVerificationCodeStatus(batchNumber, code),
            "AntiCounterfeiting: the verification code is error status"
        );
        require(
            false == attribute[id].isUsed,
            "AntiCounterfeiting: the AntiCounterfeiting token is already been used"
        );

        // 进行转账
        uint256[] memory ids = new uint256[](1);
        ids[0] = id;
        uint256[] memory amounts = new uint256[](1);
        amounts[0] = amount;
        batchTransferWithEvent(from, to, ids, amounts, data);

        // 转账完成后，修改验证码的使用状态
        batchVerificationCode[batchNumber][code] = 2;
        attribute[id].isUsed = true;

        emit GoodsUsedResult(from, to, id, code, true);

        BatchNumberVerificationCode[]
        memory batchNumberVerificationCodeList = new BatchNumberVerificationCode[](
            1
        );
        batchNumberVerificationCodeList[0] = BatchNumberVerificationCode(
            batchNumber,
            code,
            2
        );
        emit VerificationCodeResult(batchNumberVerificationCodeList);
    }

    /**
     * 设置批次对应的验证码
     * batchNumber 批次号
     * codes 验证码数组
     */
    function setVerificationCode(
        string memory batchNumber,
        string[] memory codes
    ) public {
        // 判断管理员才能操作
        require(
            _msgSender() == contractDeployer,
            "AntiCounterfeiting: setVerificationCode must only be called by the contractDeployer"
        );

        BatchNumberVerificationCode[]
        memory batchNumberVerificationCodeList = new BatchNumberVerificationCode[](
            codes.length
        );
        for (uint256 i = 0; i < codes.length; ++i) {
            batchVerificationCode[batchNumber][codes[i]] = 1;
            batchNumberVerificationCodeList[i] = BatchNumberVerificationCode(
                batchNumber,
                codes[i],
                2
            );
        }

        emit VerificationCodeResult(batchNumberVerificationCodeList);
    }

    /**
     * 查询批次对应的验证码
     * batchNumber 批次号
     * code 验证码
     *
     * 返回参数
     * status 验证码状态，0:验证码与批次未关联，1:验证码未使用, 2:验证码已使用
     */
    function getVerificationCodeStatus(
        string memory batchNumber,
        string memory code
    ) public view returns (uint8 status) {
        // 查询验证码状态
        return batchVerificationCode[batchNumber][code];
    }

    /**
     * 查询批次对应的验证码
     * goodsID 币的id，根据币id查询到币id关联的批次号
     * code 验证码
     *
     * 返回参数
     * status 验证码状态，0:验证码与批次未关联，1:验证码未使用, 2:验证码已使用
     */
    function getVerificationCodeStatusByGoodsID(
        uint256 goodsID,
        string memory code
    ) public view returns (uint8 status) {
        // 查询验证码状态
        string memory batchNumber = goodsBatchNumber[goodsID];
        return batchVerificationCode[batchNumber][code];
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
            "AntiCounterfeiting: setMintWhiteList must only be called by the contractDeployer"
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
