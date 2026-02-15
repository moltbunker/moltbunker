// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/BunkerToken.sol";

contract BunkerTokenTest is Test {
    BunkerToken public token;

    address public owner = makeAddr("owner");
    address public alice = makeAddr("alice");
    address public bob = makeAddr("bob");

    uint256 public constant SUPPLY_CAP = 100_000_000_000 * 1e18;

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);
    event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner);
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);

    function setUp() public {
        token = new BunkerToken(owner);
    }

    // -----------------------------------------------------------------------
    //  1. Deployment Tests
    // -----------------------------------------------------------------------

    function test_Deployment_Name() public view {
        assertEq(token.name(), "Bunker Token");
    }

    function test_Deployment_Symbol() public view {
        assertEq(token.symbol(), "BUNKER");
    }

    function test_Deployment_Decimals() public view {
        assertEq(token.decimals(), 18);
    }

    function test_Deployment_InitialSupplyIsZero() public view {
        assertEq(token.totalSupply(), 0);
    }

    function test_Deployment_OwnerIsCorrect() public view {
        assertEq(token.owner(), owner);
    }

    function test_Deployment_VersionIs100() public view {
        assertEq(token.VERSION(), "1.0.0");
    }

    function test_Deployment_SupplyCapConstant() public view {
        assertEq(token.SUPPLY_CAP(), SUPPLY_CAP);
    }

    function test_Deployment_MintableSupplyStartsAtCap() public view {
        assertEq(token.mintableSupply(), SUPPLY_CAP);
    }

    function test_Deployment_ZeroAddressOwnerReverts() public {
        vm.expectRevert(abi.encodeWithSignature("OwnableInvalidOwner(address)", address(0)));
        new BunkerToken(address(0));
    }

    // -----------------------------------------------------------------------
    //  2. Minting Tests
    // -----------------------------------------------------------------------

    function test_Mint_OwnerCanMintToAnyAddress() public {
        vm.prank(owner);
        token.mint(alice, 1000e18);

        assertEq(token.balanceOf(alice), 1000e18);
        assertEq(token.totalSupply(), 1000e18);
    }

    function test_Mint_OwnerCanMintToSelf() public {
        vm.prank(owner);
        token.mint(owner, 500e18);

        assertEq(token.balanceOf(owner), 500e18);
    }

    function test_Mint_EmitsTransferEvent() public {
        vm.prank(owner);
        vm.expectEmit(true, true, true, true);
        emit Transfer(address(0), alice, 1000e18);
        token.mint(alice, 1000e18);
    }

    function test_Mint_NonOwnerCannotMint() public {
        vm.prank(alice);
        vm.expectRevert(
            abi.encodeWithSignature("OwnableUnauthorizedAccount(address)", alice)
        );
        token.mint(alice, 1000e18);
    }

    function test_Mint_ZeroAmountReverts() public {
        vm.prank(owner);
        vm.expectRevert(BunkerToken.MintAmountZero.selector);
        token.mint(alice, 0);
    }

    function test_Mint_CanMintExactlyUpToCap() public {
        vm.prank(owner);
        token.mint(alice, SUPPLY_CAP);

        assertEq(token.totalSupply(), SUPPLY_CAP);
        assertEq(token.balanceOf(alice), SUPPLY_CAP);
        assertEq(token.mintableSupply(), 0);
    }

    function test_Mint_BeyondCapReverts() public {
        vm.prank(owner);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerToken.SupplyCapExceeded.selector,
                SUPPLY_CAP + 1,
                SUPPLY_CAP
            )
        );
        token.mint(alice, SUPPLY_CAP + 1);
    }

    function test_Mint_BeyondCapAfterPartialMintReverts() public {
        vm.startPrank(owner);
        token.mint(alice, SUPPLY_CAP - 100e18);

        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerToken.SupplyCapExceeded.selector,
                200e18,
                100e18
            )
        );
        token.mint(alice, 200e18);
        vm.stopPrank();
    }

    function test_Mint_MintableSupplyDecreases() public {
        vm.prank(owner);
        token.mint(alice, 300e18);

        assertEq(token.mintableSupply(), SUPPLY_CAP - 300e18);
    }

    function test_Mint_MultipleMints() public {
        vm.startPrank(owner);
        token.mint(alice, 100e18);
        token.mint(bob, 200e18);
        token.mint(alice, 50e18);
        vm.stopPrank();

        assertEq(token.balanceOf(alice), 150e18);
        assertEq(token.balanceOf(bob), 200e18);
        assertEq(token.totalSupply(), 350e18);
        assertEq(token.mintableSupply(), SUPPLY_CAP - 350e18);
    }

    function test_Mint_CanMintRemainingAfterPartialMint() public {
        vm.startPrank(owner);
        token.mint(alice, SUPPLY_CAP - 1);
        token.mint(bob, 1);
        vm.stopPrank();

        assertEq(token.totalSupply(), SUPPLY_CAP);
        assertEq(token.mintableSupply(), 0);
    }

    function test_Mint_AtCapZeroMintableReverts() public {
        vm.startPrank(owner);
        token.mint(alice, SUPPLY_CAP);

        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerToken.SupplyCapExceeded.selector,
                1,
                0
            )
        );
        token.mint(alice, 1);
        vm.stopPrank();
    }

    // -----------------------------------------------------------------------
    //  3. Burning Tests
    // -----------------------------------------------------------------------

    function test_Burn_HolderCanBurnOwnTokens() public {
        vm.prank(owner);
        token.mint(alice, 1000e18);

        vm.prank(alice);
        token.burn(400e18);

        assertEq(token.balanceOf(alice), 600e18);
    }

    function test_Burn_ReducesTotalSupply() public {
        vm.prank(owner);
        token.mint(alice, 1000e18);

        vm.prank(alice);
        token.burn(400e18);

        assertEq(token.totalSupply(), 600e18);
    }

    function test_Burn_EmitsTransferEventToZeroAddress() public {
        vm.prank(owner);
        token.mint(alice, 1000e18);

        vm.prank(alice);
        vm.expectEmit(true, true, true, true);
        emit Transfer(alice, address(0), 400e18);
        token.burn(400e18);
    }

    function test_Burn_IncreasesMintableSupply() public {
        vm.prank(owner);
        token.mint(alice, 1000e18);

        assertEq(token.mintableSupply(), SUPPLY_CAP - 1000e18);

        vm.prank(alice);
        token.burn(400e18);

        assertEq(token.mintableSupply(), SUPPLY_CAP - 600e18);
    }

    function test_Burn_CanRemintBurnedTokens() public {
        vm.prank(owner);
        token.mint(alice, SUPPLY_CAP);

        vm.prank(alice);
        token.burn(500e18);

        vm.prank(owner);
        token.mint(bob, 500e18);

        assertEq(token.totalSupply(), SUPPLY_CAP);
        assertEq(token.mintableSupply(), 0);
    }

    function test_BurnFrom_WorksWithApproval() public {
        vm.prank(owner);
        token.mint(alice, 1000e18);

        vm.prank(alice);
        token.approve(bob, 500e18);

        vm.prank(bob);
        token.burnFrom(alice, 300e18);

        assertEq(token.balanceOf(alice), 700e18);
        assertEq(token.allowance(alice, bob), 200e18);
        assertEq(token.totalSupply(), 700e18);
    }

    function test_BurnFrom_WithoutApprovalReverts() public {
        vm.prank(owner);
        token.mint(alice, 1000e18);

        vm.prank(bob);
        vm.expectRevert(
            abi.encodeWithSignature(
                "ERC20InsufficientAllowance(address,uint256,uint256)",
                bob,
                0,
                500e18
            )
        );
        token.burnFrom(alice, 500e18);
    }

    function test_BurnFrom_ExceedingApprovalReverts() public {
        vm.prank(owner);
        token.mint(alice, 1000e18);

        vm.prank(alice);
        token.approve(bob, 200e18);

        vm.prank(bob);
        vm.expectRevert(
            abi.encodeWithSignature(
                "ERC20InsufficientAllowance(address,uint256,uint256)",
                bob,
                200e18,
                500e18
            )
        );
        token.burnFrom(alice, 500e18);
    }

    function test_Burn_ExceedingBalanceReverts() public {
        vm.prank(owner);
        token.mint(alice, 100e18);

        vm.prank(alice);
        vm.expectRevert(
            abi.encodeWithSignature(
                "ERC20InsufficientBalance(address,uint256,uint256)",
                alice,
                100e18,
                200e18
            )
        );
        token.burn(200e18);
    }

    // -----------------------------------------------------------------------
    //  4. Transfer Tests
    // -----------------------------------------------------------------------

    function test_Transfer_Works() public {
        vm.prank(owner);
        token.mint(alice, 1000e18);

        vm.prank(alice);
        token.transfer(bob, 300e18);

        assertEq(token.balanceOf(alice), 700e18);
        assertEq(token.balanceOf(bob), 300e18);
    }

    function test_Transfer_EmitsEvent() public {
        vm.prank(owner);
        token.mint(alice, 1000e18);

        vm.prank(alice);
        vm.expectEmit(true, true, true, true);
        emit Transfer(alice, bob, 300e18);
        token.transfer(bob, 300e18);
    }

    function test_Transfer_ToZeroAddressReverts() public {
        vm.prank(owner);
        token.mint(alice, 1000e18);

        vm.prank(alice);
        vm.expectRevert(
            abi.encodeWithSignature("ERC20InvalidReceiver(address)", address(0))
        );
        token.transfer(address(0), 100e18);
    }

    function test_Transfer_ExceedingBalanceReverts() public {
        vm.prank(owner);
        token.mint(alice, 100e18);

        vm.prank(alice);
        vm.expectRevert(
            abi.encodeWithSignature(
                "ERC20InsufficientBalance(address,uint256,uint256)",
                alice,
                100e18,
                200e18
            )
        );
        token.transfer(bob, 200e18);
    }

    function test_Transfer_ZeroAmountWorks() public {
        vm.prank(alice);
        token.transfer(bob, 0);

        assertEq(token.balanceOf(alice), 0);
        assertEq(token.balanceOf(bob), 0);
    }

    function test_Approve_And_TransferFrom() public {
        vm.prank(owner);
        token.mint(alice, 1000e18);

        vm.prank(alice);
        token.approve(bob, 500e18);
        assertEq(token.allowance(alice, bob), 500e18);

        vm.prank(bob);
        token.transferFrom(alice, bob, 300e18);

        assertEq(token.balanceOf(alice), 700e18);
        assertEq(token.balanceOf(bob), 300e18);
        assertEq(token.allowance(alice, bob), 200e18);
    }

    function test_TransferFrom_WithoutApprovalReverts() public {
        vm.prank(owner);
        token.mint(alice, 1000e18);

        vm.prank(bob);
        vm.expectRevert(
            abi.encodeWithSignature(
                "ERC20InsufficientAllowance(address,uint256,uint256)",
                bob,
                0,
                500e18
            )
        );
        token.transferFrom(alice, bob, 500e18);
    }

    function test_TransferFrom_ExceedingApprovalReverts() public {
        vm.prank(owner);
        token.mint(alice, 1000e18);

        vm.prank(alice);
        token.approve(bob, 200e18);

        vm.prank(bob);
        vm.expectRevert(
            abi.encodeWithSignature(
                "ERC20InsufficientAllowance(address,uint256,uint256)",
                bob,
                200e18,
                500e18
            )
        );
        token.transferFrom(alice, bob, 500e18);
    }

    function test_Approve_EmitsEvent() public {
        vm.prank(alice);
        vm.expectEmit(true, true, true, true);
        emit Approval(alice, bob, 500e18);
        token.approve(bob, 500e18);
    }

    function test_Approve_MaxAllowanceDoesNotDecrease() public {
        vm.prank(owner);
        token.mint(alice, 1000e18);

        vm.prank(alice);
        token.approve(bob, type(uint256).max);

        vm.prank(bob);
        token.transferFrom(alice, bob, 500e18);

        // Infinite approval should not decrease
        assertEq(token.allowance(alice, bob), type(uint256).max);
    }

    // -----------------------------------------------------------------------
    //  5. Ownership Tests (Ownable2Step)
    // -----------------------------------------------------------------------

    function test_Ownership_TransferIsPendingUntilAccepted() public {
        vm.prank(owner);
        token.transferOwnership(alice);

        // Owner has NOT changed yet
        assertEq(token.owner(), owner);
        assertEq(token.pendingOwner(), alice);
    }

    function test_Ownership_AcceptOwnershipCompletesTransfer() public {
        vm.prank(owner);
        token.transferOwnership(alice);

        vm.prank(alice);
        token.acceptOwnership();

        assertEq(token.owner(), alice);
        assertEq(token.pendingOwner(), address(0));
    }

    function test_Ownership_EmitsEventsOnTransfer() public {
        vm.prank(owner);
        vm.expectEmit(true, true, false, false);
        emit OwnershipTransferStarted(owner, alice);
        token.transferOwnership(alice);

        vm.prank(alice);
        vm.expectEmit(true, true, false, false);
        emit OwnershipTransferred(owner, alice);
        token.acceptOwnership();
    }

    function test_Ownership_OnlyPendingOwnerCanAccept() public {
        vm.prank(owner);
        token.transferOwnership(alice);

        vm.prank(bob);
        vm.expectRevert(
            abi.encodeWithSignature("OwnableUnauthorizedAccount(address)", bob)
        );
        token.acceptOwnership();
    }

    function test_Ownership_OldOwnerLosesMintPrivilegeAfterTransfer() public {
        vm.prank(owner);
        token.transferOwnership(alice);

        vm.prank(alice);
        token.acceptOwnership();

        // Old owner can no longer mint
        vm.prank(owner);
        vm.expectRevert(
            abi.encodeWithSignature("OwnableUnauthorizedAccount(address)", owner)
        );
        token.mint(owner, 100e18);

        // New owner can mint
        vm.prank(alice);
        token.mint(alice, 100e18);
        assertEq(token.balanceOf(alice), 100e18);
    }

    function test_Ownership_NonOwnerCannotInitiateTransfer() public {
        vm.prank(alice);
        vm.expectRevert(
            abi.encodeWithSignature("OwnableUnauthorizedAccount(address)", alice)
        );
        token.transferOwnership(bob);
    }

    function test_Ownership_CanCancelPendingTransfer() public {
        vm.startPrank(owner);
        token.transferOwnership(alice);
        assertEq(token.pendingOwner(), alice);

        // Transferring to a new address cancels the pending one
        token.transferOwnership(bob);
        assertEq(token.pendingOwner(), bob);
        vm.stopPrank();

        // Alice can no longer accept
        vm.prank(alice);
        vm.expectRevert(
            abi.encodeWithSignature("OwnableUnauthorizedAccount(address)", alice)
        );
        token.acceptOwnership();

        // Bob can accept
        vm.prank(bob);
        token.acceptOwnership();
        assertEq(token.owner(), bob);
    }

    function test_Ownership_RenounceOwnership() public {
        vm.prank(owner);
        token.renounceOwnership();

        assertEq(token.owner(), address(0));
    }

    function test_Ownership_RenouncePreventsMinting() public {
        vm.prank(owner);
        token.renounceOwnership();

        vm.prank(owner);
        vm.expectRevert(
            abi.encodeWithSignature("OwnableUnauthorizedAccount(address)", owner)
        );
        token.mint(alice, 100e18);
    }

    // -----------------------------------------------------------------------
    //  6. Permit (EIP-2612) Tests
    // -----------------------------------------------------------------------

    function test_Permit_ValidSignature() public {
        uint256 alicePrivateKey = 0xa11ce;
        address aliceSigner = vm.addr(alicePrivateKey);

        vm.prank(owner);
        token.mint(aliceSigner, 1000e18);

        uint256 nonce = token.nonces(aliceSigner);
        uint256 deadline = block.timestamp + 1 hours;
        uint256 amount = 500e18;

        bytes32 permitHash = _getPermitHash(aliceSigner, bob, amount, nonce, deadline);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(alicePrivateKey, permitHash);

        token.permit(aliceSigner, bob, amount, deadline, v, r, s);

        assertEq(token.allowance(aliceSigner, bob), amount);
        assertEq(token.nonces(aliceSigner), 1);
    }

    function test_Permit_ExpiredDeadlineReverts() public {
        uint256 alicePrivateKey = 0xa11ce;
        address aliceSigner = vm.addr(alicePrivateKey);

        uint256 deadline = block.timestamp - 1; // Already expired
        uint256 amount = 500e18;
        uint256 nonce = token.nonces(aliceSigner);

        bytes32 permitHash = _getPermitHash(aliceSigner, bob, amount, nonce, deadline);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(alicePrivateKey, permitHash);

        vm.expectRevert(
            abi.encodeWithSignature(
                "ERC2612ExpiredSignature(uint256)",
                deadline
            )
        );
        token.permit(aliceSigner, bob, amount, deadline, v, r, s);
    }

    function test_Permit_InvalidSignatureReverts() public {
        uint256 alicePrivateKey = 0xa11ce;
        address aliceSigner = vm.addr(alicePrivateKey);
        uint256 wrongPrivateKey = 0xb0b;

        uint256 deadline = block.timestamp + 1 hours;
        uint256 amount = 500e18;
        uint256 nonce = token.nonces(aliceSigner);

        bytes32 permitHash = _getPermitHash(aliceSigner, bob, amount, nonce, deadline);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(wrongPrivateKey, permitHash);

        vm.expectRevert(
            abi.encodeWithSignature(
                "ERC2612InvalidSigner(address,address)",
                vm.addr(wrongPrivateKey),
                aliceSigner
            )
        );
        token.permit(aliceSigner, bob, amount, deadline, v, r, s);
    }

    function test_Permit_NoncesIncrementCorrectly() public {
        uint256 alicePrivateKey = 0xa11ce;
        address aliceSigner = vm.addr(alicePrivateKey);

        assertEq(token.nonces(aliceSigner), 0);

        uint256 deadline = block.timestamp + 1 hours;

        // First permit
        bytes32 hash1 = _getPermitHash(aliceSigner, bob, 100e18, 0, deadline);
        (uint8 v1, bytes32 r1, bytes32 s1) = vm.sign(alicePrivateKey, hash1);
        token.permit(aliceSigner, bob, 100e18, deadline, v1, r1, s1);
        assertEq(token.nonces(aliceSigner), 1);

        // Second permit
        bytes32 hash2 = _getPermitHash(aliceSigner, bob, 200e18, 1, deadline);
        (uint8 v2, bytes32 r2, bytes32 s2) = vm.sign(alicePrivateKey, hash2);
        token.permit(aliceSigner, bob, 200e18, deadline, v2, r2, s2);
        assertEq(token.nonces(aliceSigner), 2);

        // Allowance reflects the latest permit value
        assertEq(token.allowance(aliceSigner, bob), 200e18);
    }

    function test_Permit_ReplayWithSameNonceReverts() public {
        uint256 alicePrivateKey = 0xa11ce;
        address aliceSigner = vm.addr(alicePrivateKey);

        uint256 deadline = block.timestamp + 1 hours;
        uint256 amount = 500e18;

        bytes32 permitHash = _getPermitHash(aliceSigner, bob, amount, 0, deadline);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(alicePrivateKey, permitHash);

        // First call succeeds
        token.permit(aliceSigner, bob, amount, deadline, v, r, s);
        assertEq(token.nonces(aliceSigner), 1);

        // Replay reverts because nonce is now 1 but signature was over nonce 0.
        // The ecrecover will yield an unpredictable address, so we just assert it reverts.
        vm.expectRevert();
        token.permit(aliceSigner, bob, amount, deadline, v, r, s);
    }

    function test_Permit_AllowsTransferFrom() public {
        uint256 alicePrivateKey = 0xa11ce;
        address aliceSigner = vm.addr(alicePrivateKey);

        vm.prank(owner);
        token.mint(aliceSigner, 1000e18);

        uint256 deadline = block.timestamp + 1 hours;
        uint256 amount = 500e18;

        bytes32 permitHash = _getPermitHash(aliceSigner, bob, amount, 0, deadline);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(alicePrivateKey, permitHash);

        token.permit(aliceSigner, bob, amount, deadline, v, r, s);

        vm.prank(bob);
        token.transferFrom(aliceSigner, bob, 300e18);

        assertEq(token.balanceOf(bob), 300e18);
        assertEq(token.balanceOf(aliceSigner), 700e18);
        assertEq(token.allowance(aliceSigner, bob), 200e18);
    }

    function test_Permit_DomainSeparator() public view {
        bytes32 domainSeparator = token.DOMAIN_SEPARATOR();
        assertTrue(domainSeparator != bytes32(0));
    }

    // -----------------------------------------------------------------------
    //  Helper: Build EIP-2612 permit digest
    // -----------------------------------------------------------------------

    function _getPermitHash(
        address permitOwner,
        address spender,
        uint256 value,
        uint256 nonce,
        uint256 deadline
    ) internal view returns (bytes32) {
        bytes32 PERMIT_TYPEHASH = keccak256(
            "Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)"
        );

        bytes32 structHash = keccak256(
            abi.encode(PERMIT_TYPEHASH, permitOwner, spender, value, nonce, deadline)
        );

        return keccak256(
            abi.encodePacked("\x19\x01", token.DOMAIN_SEPARATOR(), structHash)
        );
    }

    // -----------------------------------------------------------------------
    //  7. Fuzz Tests
    // -----------------------------------------------------------------------

    /// @notice Fuzz token transfers: mint to alice, transfer a portion to bob.
    function testFuzz_transfer(uint128 amount) public {
        // Constrain amount to stay within the supply cap
        vm.assume(amount > 0 && amount <= SUPPLY_CAP);

        // Mint tokens to alice
        vm.prank(owner);
        token.mint(alice, amount);

        assertEq(token.balanceOf(alice), amount);

        // Transfer all to bob
        vm.prank(alice);
        token.transfer(bob, amount);

        assertEq(token.balanceOf(alice), 0);
        assertEq(token.balanceOf(bob), amount);
        assertEq(token.totalSupply(), amount);
    }

    /// @notice Fuzz approve + transferFrom: approve some amount, then transferFrom up to that.
    function testFuzz_approve_and_transferFrom(uint128 approveAmount, uint128 transferAmount) public {
        // Ensure we have enough tokens and transfer doesn't exceed approval
        vm.assume(approveAmount > 0);
        vm.assume(transferAmount > 0 && transferAmount <= approveAmount);
        vm.assume(uint256(approveAmount) <= SUPPLY_CAP);

        // Mint tokens to alice
        vm.prank(owner);
        token.mint(alice, approveAmount);

        // Alice approves bob
        vm.prank(alice);
        token.approve(bob, approveAmount);
        assertEq(token.allowance(alice, bob), approveAmount);

        // Bob transfers from alice
        vm.prank(bob);
        token.transferFrom(alice, bob, transferAmount);

        assertEq(token.balanceOf(alice), uint256(approveAmount) - uint256(transferAmount));
        assertEq(token.balanceOf(bob), transferAmount);
        assertEq(token.allowance(alice, bob), uint256(approveAmount) - uint256(transferAmount));
    }

    /// @notice Fuzz minting within cap: mint a random amount within the supply cap.
    function testFuzz_mint_withinCap(uint128 amount) public {
        vm.assume(amount > 0 && uint256(amount) <= SUPPLY_CAP);

        vm.prank(owner);
        token.mint(alice, amount);

        assertEq(token.balanceOf(alice), amount);
        assertEq(token.totalSupply(), amount);
        assertEq(token.mintableSupply(), SUPPLY_CAP - uint256(amount));
    }

    /// @notice Fuzz burning: mint some tokens, then burn a portion.
    function testFuzz_burn(uint128 amount) public {
        vm.assume(amount > 0 && uint256(amount) <= SUPPLY_CAP);

        // Mint tokens to alice
        vm.prank(owner);
        token.mint(alice, amount);

        // Alice burns all her tokens
        vm.prank(alice);
        token.burn(amount);

        assertEq(token.balanceOf(alice), 0);
        assertEq(token.totalSupply(), 0);
        assertEq(token.mintableSupply(), SUPPLY_CAP);
    }
}
