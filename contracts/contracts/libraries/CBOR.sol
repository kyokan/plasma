pragma solidity 0.4.18;
/**
* @title CBOReader
*
* CBOReader is used to read and parse CBOR encoded data in memory.
*
* @author mattkim (matthkim@gmail.com)
*/
library CBOR {



 struct CBORItem {
     uint _unsafe_memPtr;    // Pointer to the CBOR-encoded bytes.
     uint _unsafe_length;    // Number of bytes. This is the full length of the string.
 }

 /* CBORItem */

 /// @dev Creates a CBORItem from an array of CBOR encoded bytes.
 /// @param self The CBOR encoded bytes.
 /// @return An CBORItem
 function toCBORItem(bytes memory self) internal constant returns (CBORItem memory) {
     uint len = self.length;
     if (len == 0) {
         return CBORItem(0, 0);
     }
     uint memPtr;
     assembly {
         memPtr := add(self, 0x20)
     }
     return CBORItem(memPtr, len);
 }
}
