
pragma solidity ^0.4.17;

import "./libraries/SafeMath.sol";

contract PriorityQueue {
    event DebugUint(address sender, uint item);

    uint256[] public priorities;
    uint256 public max = 2**256-1;

    function add(uint256 priority) {
        assert(priority != SafeMath.max());
        // TODO: Throws an invalid opcode for some reason.
        // uint256 length = priorities.length + 1;
        priorities[priorities.length++] = priority;
        bubbleUp();
    }

    function getPriorities() {
        DebugUint(msg.sender, priorities.length);

        for (uint256 i = 0; i < priorities.length; i++) {
            DebugUint(msg.sender, priorities[i]);
        }
    }

    function bubbleUp() {
        uint256[] storage p = priorities;

        uint256 i = p.length - 1;

        while (i > 0) {
            // Parent
            uint256 j;

            if (i % 2 == 1) {
                j = i / 2;
            } else {
                j = i / 2 - 1;
            }

            if (p[i] < p[j]) {
                uint256 tmp = p[i];
                p[i] = p[j];
                p[j] = tmp;
            }
            else {
                break;
            }

            i = j;
        }
    }

    function remove(uint256 id) returns (bool) {
        uint256[] storage p = priorities;
        uint256 i = 0;
        while (p[i] != id && i < priorities.length) {
            i++;
        }

        // We didn't find a match.
        if (i >= priorities.length) {
            return false;
        }

        p[i] = SafeMath.max();
        bubbleDown(i);
        return true;
    }

    function pop() returns (uint256) {
        uint256[] storage p = priorities;

        if (p.length == 0) {
            return SafeMath.max();
        }

        uint256 res = p[0];
        p[0] = SafeMath.max();
        bubbleDown(0);
        return res;
    }

    function bubbleDown(uint256 i) {
        uint256[] storage p = priorities;

        while(i * 2 + 1 < p.length) {
            uint256 j = i * 2 + 1;
            uint256 k = i * 2 + 2;

            uint256 parent = p[i];
            uint256 left = p[j];

            if(k >= p.length) {
                p[i] = left;
                p[j] = parent;
                break;
            }

            uint256 right = p[k];

            if (left < right) {
                p[i] = left;
                p[j] = parent;
                i = j;
            }
            else {
                // If we're equal and both are maxes
                // Then we move right which makes the
                // maxes right heavy.
                p[i] = right;
                p[k] = parent;
                i = k;
            }
        }

        prune();
    }

    function prune() {
        uint256[] storage p = priorities;
        uint256 i = p.length - 1;

        while(i > 0 && p[i] == SafeMath.max()) {
            p.length--;
            i--;
        }

        if (i == 0 && p[i] == SafeMath.max()) {
            p.length = 0;
        }
    }
}