// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_frontend/api_service.dart';
import 'package:flutter_frontend/resources/app_style.dart';
import 'package:flutter_frontend/views/auth/auth_controller.dart';
import 'package:provider/provider.dart';

class WalletCard extends StatefulWidget {
  const WalletCard({
    super.key,
  });

  @override
  State<WalletCard> createState() => _WalletCardState();
}

class _WalletCardState extends State<WalletCard> {
  final _addFundsTextController = TextEditingController();
  int currWallet = 0;

  @override
  void dispose() {
    _addFundsTextController.dispose();
    super.dispose();
  }

  // ...this feels weird? Having the function be definied WITHIN another function, rather than root of class.
  void addMoney(APIService apiService) async {
    if (_addFundsTextController.text.isEmpty) {
      return;
    }

    int amount = int.parse(_addFundsTextController.text);

    Response response = await apiService.addMoneyToWallet(amount);
    final data = response.data;

    if (data is Map && data.containsKey('success') && data['success'] == true) {
      // This is TECHNICALLY Working I THINK. The issue is that getWallet() API call is made when the builder
      //    gets rebuilt. Since I'm mocking, that always returns 100 and overwrites this change.
      // Actually IT WORKS! I tested by changing the getWallet() mock to accept the curr wallet and return that. Sweet!
      setState(() {
        currWallet = currWallet + amount;
        print("Updating the wallet to $currWallet");
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final APIService apiService = APIService(
      Provider.of<AuthController>(context, listen: false),
    );

    return Expanded(
      child: FutureBuilder<void>(
        future: apiService.getWalletBalance(),
        builder: (context, snapshot) {
          if (snapshot.connectionState == ConnectionState.waiting) {
            return Center(
              child: const CircularProgressIndicator(),
            );
          }
          else if (snapshot.hasError) {
            print('>> Connection error: ${snapshot.error}');
            return Center(
              child: const Text(
                'Something has gone terribly wrong - connection error.',
                style: MyAppStyle.largeFont,
              ),
            );
          }
          else if (snapshot.connectionState == ConnectionState.done) {
            final Response response = snapshot.data as Response;
            final data = response.data;

            if (data is Map && data.containsKey('success') && data['success'] == true) {
              currWallet = data['data'][0]['balance'];

              return Card(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.center,
                  children: [
                    Text(
                      'Wallet',
                      style: MyAppStyle.largeFont
                    ),
                    SizedBox(
                      height: 50,
                    ),
                    Text(
                      'Current Total: \$$currWallet',
                      style: MyAppStyle.regularFont,
                    ),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Text(
                          'Add More?',
                          style: MyAppStyle.regularFont,
                        ),
                        IconButton(
                          onPressed: () => {
                            addMoney(apiService),
                          },
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
                    ),
                  ],
                ),
              );
            }
            else {
              print(">> Unexpected response behaviour.");
              return Column(
                children: [
                  Text(
                    'Wallet',
                    style: MyAppStyle.largeFont,
                  ),
                  Text(
                    'Unexpected network error.',
                    style: MyAppStyle.regularFontLightGrey,
                  )
                ],
              );
            }
          }
          else {
            return Center(
              child: const Text(
                'Something has gone terribly wrong - Unhanddled connection state.',
                style: MyAppStyle.largeFont,
              ),
            );
          }
        },
      ),
    );
  }
}