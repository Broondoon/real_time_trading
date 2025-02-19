// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_frontend/resources/app_style.dart';

class WalletCard extends StatefulWidget {
  const WalletCard({
    super.key,
  });

  @override
  State<WalletCard> createState() => _WalletCardState();
}

class _WalletCardState extends State<WalletCard> {
  final _addFundsTextController = TextEditingController();

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Card(
        child: Center(
          child: Column(
            children: [
              Text(
                'Wallet',
                style: MyAppStyle.largeFont
              ),
              Row(
                children: [
                  Text(
                    'Current Total:',
                    style: MyAppStyle.regularFont,
                  ),
                  Text(
                    '\$99999.99',
                    style: MyAppStyle.regularFont,
                  ),
                ],
              ),
              Row(
                children: [
                  Text(
                    'Add More?',
                    style: MyAppStyle.regularFont,
                  ),
                  IconButton(
                    onPressed: () => {},
                    icon: Icon(
                      Icons.add,
                    ),
                  ),
                  SizedBox(
                    height: 50,
                    width: 100,
                    child: Card(
                      child: Padding(
                        padding: const EdgeInsets.all(8.0),
                        child: TextField(
                          controller: _addFundsTextController,
                          maxLength: 7,
                          decoration: const InputDecoration(
                            // labelText: 'Funds',
                            // labelStyle: MyAppStyle.regularFont,
                            hintText: '\$000.00',
                            hintStyle: MyAppStyle.regularFontLightGrey,
                            border: InputBorder.none,
                            counterText: '',
                          ),
                          keyboardType: TextInputType.number,
                          inputFormatters: [
                            FilteringTextInputFormatter.digitsOnly,
                          ],
                        ),
                      ),
                    ),
                  ),
                ],
              )
            ],
          )
        ),
      ),
    );
  }
}
